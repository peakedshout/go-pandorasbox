package ticker

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type HandleFunc func(id string) error

type Ticker struct {
	m sync.Map

	delay atomic.Pointer[time.Duration]

	count atomic.Uint64

	ctx context.Context
	cl  context.CancelFunc
}

type unit struct {
	id   string
	lock sync.Mutex

	last time.Time
	c    chan time.Duration

	ctx context.Context
}

func NewTicker(ctx context.Context) *Ticker {
	if ctx == nil {
		ctx = context.Background()
	}
	t := &Ticker{}
	t.delay.Store(new(time.Duration))
	t.ctx, t.cl = context.WithCancel(ctx)
	return t
}

func (t *Ticker) Stop() {
	t.cl()
}

// GetDelay The last valid data will be fetched, the initialized data is 0, and invalid data will always be ignored
func (t *Ticker) GetDelay() time.Duration {
	return *t.delay.Load()
}

// DelayOnce
//
//	Output -1 if startup fails and 0 if ctx shuts down or times out
func (t *Ticker) DelayOnce(ctx context.Context, fn HandleFunc) time.Duration {
	if ctx == nil {
		var cl context.CancelFunc
		ctx, cl = context.WithCancel(context.Background())
		defer cl()
	}
	if t.disable() {
		return 0
	}
	u := t.newUnit(ctx)
	defer t.delUnit(u.id)
	u.lock.Lock()
	u.last = time.Now()
	u.lock.Unlock()
	err := fn(u.id)
	if err != nil {
		return -1
	}
	select {
	case <-t.ctx.Done():
	case <-u.ctx.Done():
	case td := <-u.c:
		t.delay.Store(&td)
		return td
	}
	return 0
}

// DelayTick
//
//	If the start failure or CTX close, close the channel
func (t *Ticker) DelayTick(ctx context.Context, interval time.Duration, fn HandleFunc) <-chan time.Duration {
	if ctx == nil {
		ctx = context.Background()
	}
	ch := make(chan time.Duration)
	if t.disable() {
		close(ch)
		return ch
	}
	if interval < 0 {
		interval = 0
	}
	go func() {
		u := t.newUnit(ctx)
		defer t.delUnit(u.id)
		defer close(ch)
		to := time.NewTimer(interval * 2)
		if !to.Stop() {
			<-to.C
		}
		defer to.Stop()
		tr := time.NewTimer(interval)
		if !tr.Stop() {
			<-tr.C
		}
		defer tr.Stop()
		var err error
		for !t.disable() {
			tr.Reset(interval)
			select {
			case <-t.ctx.Done():
				return
			case <-u.ctx.Done():
				return
			case <-tr.C:
				u.lock.Lock()
				u.last = time.Now()
				u.lock.Unlock()
				err = fn(u.id)
				if err != nil {
					return
				}
				to.Reset(2 * interval)
			}
			select {
			case <-to.C:
				td := time.Duration(-1)
				t.delay.Store(&td)
				continue
			case <-t.ctx.Done():
				return
			case <-u.ctx.Done():
				return
			case td := <-u.c:
				ch <- td
				t.delay.Store(&td)
				if !to.Stop() {
					<-to.C
				}
				continue
			}
		}
	}()
	return ch
}

// Record
//
//	If you don't accept treatment in time, it discards the time record
func (t *Ticker) Record(id string) {
	if t.disable() {
		return
	}
	u, ok := t.getUnit(id)
	if !ok {
		return
	}
	u.lock.Lock()
	defer u.lock.Unlock()
	tn := time.Now()
	td := tn.Sub(u.last)
	select {
	case u.c <- td:
	default:
	}
}

func (t *Ticker) disable() bool {
	select {
	case <-t.ctx.Done():
		return true
	default:
		return false
	}
}

func (t *Ticker) newUnit(ctx context.Context) *unit {
	u := &unit{
		id:  strconv.FormatUint(t.count.Add(1), 10),
		c:   make(chan time.Duration),
		ctx: ctx,
	}
	t.setUnit(u.id, u)
	return u
}

func (t *Ticker) getUnit(id string) (*unit, bool) {
	a, ok := t.m.Load(id)
	if !ok {
		return nil, ok
	}
	return a.(*unit), ok
}

func (t *Ticker) getAndDelUnit(id string) (*unit, bool) {
	a, ok := t.m.LoadAndDelete(id)
	if !ok {
		return nil, ok
	}
	return a.(*unit), ok
}

func (t *Ticker) delUnit(id string) {
	t.m.Delete(id)
}

func (t *Ticker) setUnit(id string, u *unit) {
	t.m.Store(id, u)
}
