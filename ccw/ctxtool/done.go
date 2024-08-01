package ctxtool

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/tool/gpool"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func Disable(ctx context.Context) bool {
	if ctx == nil {
		return true
	}
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func Wait(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func WaitFunc(ctx context.Context, fn func()) {
	<-ctx.Done()
	fn()
}

func GWaitFunc(ctx context.Context, fn func()) {
	_cwmOnce.Do(func() {
		_cwm = NewCtxWaiterManager(context.Background(), 0)
	})
	_cwm.WaitFunc(ctx, fn)
}

var _cwm *CtxWaiterManager
var _cwmOnce sync.Once

func NewCtxWaiterManager(ctx context.Context, num int) *CtxWaiterManager {
	cwm := &CtxWaiterManager{
		sleep: make(chan *waitUnit),
		num:   num,
	}
	cwm.Context, cwm.cancel = context.WithCancel(ctx)
	go cwm.run()
	return cwm
}

// CtxWaiterManager maximum of 4294705156(65534*65534)?
type CtxWaiterManager struct {
	context.Context
	cancel  context.CancelFunc
	sleep   chan *waitUnit
	waiters []*CtxWaiter
	num     int
}

func (cwm *CtxWaiterManager) Stop() {
	cwm.cancel()
}

func (cwm *CtxWaiterManager) WaitFunc(ctx context.Context, fn func()) bool {
	if ctx.Err() != nil {
		fn()
		return true
	}
	unit := &waitUnit{
		ctx: ctx,
		fn:  fn,
	}
	return cwm.waitFunc(unit)
}

func (cwm *CtxWaiterManager) run() {
	tk := time.NewTicker(30 * time.Second)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			cwm.waitFunc(nil)
		case <-cwm.Context.Done():
			return
		case unit := <-cwm.sleep:
			cwm.waitFunc(unit)
		}
	}
}

func (cwm *CtxWaiterManager) waitFunc(unit *waitUnit) bool {
	wait := cwm.toWait(unit)
	switch wait {
	case 0:
		if unit != nil {
			unit.fn()
		}
		return false
	default:
		return true
	}
}

func (cwm *CtxWaiterManager) toWait(unit *waitUnit) int {
	if cwm.Context.Err() != nil {
		return 0
	}
	if unit == nil {
		sort.Slice(cwm.waiters, func(i, j int) bool {
			return cwm.waiters[i].getCount() > cwm.waiters[j].getCount()
		})
		once := -1
		for i, waiter := range cwm.waiters {
			if waiter.getCount() == 0 {
				if once != -1 {
					waiter.Stop()
					continue
				} else {
					once = i
				}
			}
		}
		if once != -1 {
			cwm.waiters = cwm.waiters[:once]
		}
		return -1
	}
	cases := make([]reflect.SelectCase, 0, len(cwm.waiters)+1)
	cases = append(cases, toCases(cwm.Context.Done()))
	sort.Slice(cwm.waiters, func(i, j int) bool {
		return cwm.waiters[i].getCount() > cwm.waiters[j].getCount()
	})
	for _, one := range cwm.waiters {
		l := one.getCount()
		if l >= 65536-3 {
			continue
		}
		cases = append(cases, cwm.toCases(one, unit))
		if l < 60000 {
			break
		}
	}
	if len(cases) != 1 {
		chosen, _, _ := reflect.Select(cases)
		return chosen
	} else {
		cwm.newCtxWaiter(unit)
		return 1
	}
}

func (cwm *CtxWaiterManager) newCtxWaiter(unit *waitUnit) {
	waiter := NewCtxWaiter(cwm.Context, cwm.num)
	waiter.syncChan = cwm.sleep
	cwm.waiters = append(cwm.waiters, waiter)
	waiter.waitFunc(unit)
}

func (cwm *CtxWaiterManager) toCases(cw *CtxWaiter, unit *waitUnit) reflect.SelectCase {
	return reflect.SelectCase{Dir: reflect.SelectSend, Chan: reflect.ValueOf(cw.sleep), Send: reflect.ValueOf(unit)}
}

func NewCtxWaiter(ctx context.Context, num int) *CtxWaiter {
	if num == 0 {
		num = runtime.NumCPU()
		if num == 1 {
			num++
		}
	}
	cw := &CtxWaiter{
		sleep: make(chan *waitUnit),
		g:     gpool.NewGPool(context.Background(), num),
	}
	cw.Context, cw.cancel = context.WithCancel(ctx)
	go cw.run()
	return cw
}

// CtxWaiter maximum of  65534(65536-2)
type CtxWaiter struct {
	context.Context
	cancel   context.CancelFunc
	sleep    chan *waitUnit
	syncChan chan *waitUnit
	g        *gpool.GPool
	ctxs     []*waitUnit
	lens     uint32
}

type waitUnit struct {
	ctx context.Context
	fn  func()
}

func (cw *CtxWaiter) Stop() {
	cw.cancel()
}

func (cw *CtxWaiter) WaitFunc(ctx context.Context, fn func()) bool {
	if ctx.Err() != nil {
		fn()
		return true
	}
	unit := &waitUnit{
		ctx: ctx,
		fn:  fn,
	}
	return cw.waitFunc(unit)
}

func (cw *CtxWaiter) waitFunc(unit *waitUnit) bool {
	select {
	case cw.sleep <- unit:
		return true
	case <-cw.Context.Done():
		unit.fn()
		return false
	}
}

func (cw *CtxWaiter) run() {
	defer func() {
		for _, ctx := range cw.ctxs {
			cw.g.Do(ctx.fn)
		}
		cw.g.Stop()
	}()
	for {
		wait, unit := cw.toWait()
		switch wait {
		case 0:
			return
		case 1:
			if len(cw.ctxs) >= 65536-2 {
				if cw.syncChan != nil {
					select {
					case <-cw.Context.Done():
						unit.fn()
					case cw.syncChan <- unit:
					}
				} else {
					cw.g.Do(unit.fn)
				}
				continue
			}
			cw.ctxs = append(cw.ctxs, unit)
		default:
			u := cw.ctxs[wait-2]
			cw.g.Do(u.fn)
			cw.ctxs = append(cw.ctxs[:wait-2], cw.ctxs[wait-1:]...)
		}
		cw.setCount()
	}
}

func (cw *CtxWaiter) setCount() {
	atomic.StoreUint32(&cw.lens, uint32(len(cw.ctxs)))
}

func (cw *CtxWaiter) getCount() int {
	return int(atomic.LoadUint32(&cw.lens))
}

func (cw *CtxWaiter) toWait() (int, *waitUnit) {
	cases := make([]reflect.SelectCase, 0, len(cw.ctxs)+2)
	cases = append(cases, toCases(cw.Context.Done()), cw.toCases(cw.sleep))
	for i := range cw.ctxs {
		cases = append(cases, toCases(cw.ctxs[i].ctx.Done()))
	}
	var unit *waitUnit
	chosen, recv, _ := reflect.Select(cases)
	if chosen == 1 {
		unit = recv.Interface().(*waitUnit)
	}
	return chosen, unit
}

func (cw *CtxWaiter) toCases(ch <-chan *waitUnit) reflect.SelectCase {
	return reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
}
