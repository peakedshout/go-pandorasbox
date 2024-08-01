package expired

import (
	"container/heap"
	"context"
	"github.com/peakedshout/go-pandorasbox/tool/gpool"
	"sync"
	"sync/atomic"
	"time"
)

type Expired interface {
	Id() any
	ExpiredFunc()
}

type ExpiredCtx struct {
	heap   expiredHeap
	lock   sync.Mutex
	closed atomic.Bool
	gp     *gpool.GPool

	topTime time.Time

	sleep  chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

func Init(ctx context.Context, num int) *ExpiredCtx {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	ec := &ExpiredCtx{
		heap:    expiredHeap{},
		gp:      gpool.NewGPool(ctx, num),
		topTime: time.Time{},
		sleep:   make(chan struct{}, 1),
		ctx:     ctx,
		cancel:  cancel,
	}
	heap.Init(&ec.heap)
	go ec.run()
	return ec
}

func (ec *ExpiredCtx) SetWithTime(a Expired, t time.Time) bool {
	ec.lock.Lock()
	defer ec.lock.Unlock()
	if ec.closed.Load() {
		return false
	}
	if time.Now().Sub(t) >= 0 {
		return ec.expired(a)
	}
	heap.Push(&ec.heap, expiredHeapUnit{
		value:       a,
		expiredTime: t,
	})
	if ec.topTime.IsZero() || ec.topTime.Sub(t) > 0 {
		ec.active()
	}
	return true
}

func (ec *ExpiredCtx) SetWithDuration(a Expired, d time.Duration) bool {
	return ec.SetWithTime(a, time.Now().Add(d))
}

func (ec *ExpiredCtx) Remove(id any, expired bool) bool {
	ec.lock.Lock()
	defer ec.lock.Unlock()
	if ec.closed.Load() {
		return false
	}
	index := -1
	for i, one := range ec.heap {
		if one.value.Id() == id {
			index = i
			break
		}
	}
	if index == -1 {
		return false
	}
	e := heap.Remove(&ec.heap, index).(expiredHeapUnit).value
	if expired {
		ec.expired(e)
	}
	ec.active()
	return true
}

func (ec *ExpiredCtx) UpdateWithTime(id any, t time.Time) bool {
	ec.lock.Lock()
	defer ec.lock.Unlock()
	if ec.closed.Load() {
		return false
	}
	index := -1
	for i, one := range ec.heap {
		if one.value.Id() == id {
			ec.heap[i].expiredTime = t
			index = i
			break
		}
	}
	if index == -1 {
		return false
	}
	heap.Fix(&ec.heap, index)
	ec.active()
	return true
}

func (ec *ExpiredCtx) UpdateWithDuration(id any, d time.Duration) bool {
	return ec.UpdateWithTime(id, time.Now().Add(d))
}

func (ec *ExpiredCtx) Stop() {
	ec.lock.Lock()
	defer ec.lock.Unlock()
	ec.once.Do(func() {
		ec.cancel()
		ec.closed.Store(true)
		ec.clear()
		ec.gp.Stop()
	})
}

func (ec *ExpiredCtx) Wait() {
	<-ec.ctx.Done()
}

func (ec *ExpiredCtx) run() {
	defer func() {
		ec.Stop()
	}()
	tk := time.NewTimer(0)
	if !tk.Stop() {
		<-tk.C
	}
	for {
		select {
		case <-ec.sleep:
			top, _, ok := ec.top()
			if !ok {
				break
			}
			td := top.Sub(time.Now())
			tk.Reset(td)
			select {
			case <-tk.C:
				a := ec.get()
				if !ec.expired(a) {
					return
				}
				ec.active()
			case <-ec.sleep:
				if !tk.Stop() {
					<-tk.C
				}
				ec.active()
			case <-ec.ctx.Done():
				return
			}
		case <-ec.ctx.Done():
			return
		}
	}
}

func (ec *ExpiredCtx) get() Expired {
	ec.lock.Lock()
	defer ec.lock.Unlock()
	a := heap.Pop(&ec.heap)
	return a.(expiredHeapUnit).value.(Expired)
}

func (ec *ExpiredCtx) top() (time.Time, Expired, bool) {
	ec.lock.Lock()
	defer ec.lock.Unlock()
	if len(ec.heap) == 0 {
		ec.topTime = time.Time{}
		return time.Time{}, nil, false
	}
	ec.topTime = ec.heap[0].expiredTime
	return ec.heap[0].expiredTime, ec.heap[0].value.(Expired), true
}

func (ec *ExpiredCtx) clear() {
	for _, one := range ec.heap {
		ec.expired(one.value)
	}
}

func (ec *ExpiredCtx) active() {
	select {
	case ec.sleep <- struct{}{}:
	default:
	}
}

func (ec *ExpiredCtx) expired(e Expired) bool {
	return ec.gp.Do(e.ExpiredFunc)
}

type expiredHeap []expiredHeapUnit

type expiredHeapUnit struct {
	value       Expired
	expiredTime time.Time
}

func (eh expiredHeap) Len() int {
	return len(eh)
}

func (eh expiredHeap) Less(i, j int) bool {
	return eh[i].expiredTime.Sub(eh[j].expiredTime) < 0
}

func (eh expiredHeap) Swap(i, j int) {
	eh[i], eh[j] = eh[j], eh[i]
}

func (eh *expiredHeap) Push(a any) {
	*eh = append(*eh, a.(expiredHeapUnit))
}

func (eh *expiredHeap) Pop() any {
	res := (*eh)[len(*eh)-1]
	*eh = (*eh)[:len(*eh)-1]
	return res
}
