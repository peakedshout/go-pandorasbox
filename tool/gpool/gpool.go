package gpool

import (
	"context"
	"sync"
)

type GPool struct {
	tasks  chan func()
	twg    sync.WaitGroup
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewGPool(ctx context.Context, num int) *GPool {
	if num < 1 {
		num = 1
	}
	ctx, cancel := context.WithCancel(ctx)
	pool := &GPool{
		tasks:  make(chan func()),
		ctx:    ctx,
		cancel: cancel,
	}
	pool.wg.Add(num)
	for i := 0; i < num; i++ {
		go pool.worker()
	}
	return pool
}

func (p *GPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case task := <-p.tasks:
			p.twg.Add(1)
			task()
			p.twg.Done()
		}
	}
}

func (p *GPool) Do(task func()) bool {
	select {
	case <-p.ctx.Done():
		task()
		return false
	case p.tasks <- task:
		return true
	}
}

func (p *GPool) Stop() {
	p.cancel()
	p.wg.Wait()
}

func (p *GPool) Wait() {
	p.twg.Wait()
}
