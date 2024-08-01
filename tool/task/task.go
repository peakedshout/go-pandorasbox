package task

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrTaskTimeout = errors.New("task timeout")
)

func NewTaskCtx[T any](ctx context.Context) *TaskCtx[T] {
	if ctx == nil {
		ctx = context.Background()
	}
	tc := &TaskCtx[T]{taskMap: make(map[any]*task[T])}
	tc.ctx, tc.cancel = context.WithCancel(ctx)
	return tc
}

type TaskCtx[T any] struct {
	ctx      context.Context
	cancel   context.CancelFunc
	taskLock sync.Mutex
	taskMap  map[any]*task[T]
}

func (tc *TaskCtx[T]) Stop() {
	tc.cancel()
}

func (tc *TaskCtx[T]) RegisterTask(ctx context.Context, id any, runner Runner) (msg *T, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	t := &task[T]{
		tctx:      tc,
		id:        id,
		initiator: runner,
		lock:      sync.Mutex{},
		waitCh:    make(chan *msgT[T]),
	}
	t.ctx, t.cancel = context.WithCancel(tc.ctx)
	defer t.cancel()
	return t.waitCtx(ctx)
}

func (tc *TaskCtx[T]) RegisterTaskWithDr(dr time.Duration, id any, runner Runner) (msg *T, err error) {
	t := &task[T]{
		tctx:      tc,
		id:        id,
		initiator: runner,
		lock:      sync.Mutex{},
		waitCh:    make(chan *msgT[T]),
	}
	t.ctx, t.cancel = context.WithCancel(tc.ctx)
	defer t.cancel()
	return t.wait(dr)
}

func (tc *TaskCtx[T]) CallBack(id any, msg *T, err error) bool {
	tc.taskLock.Lock()
	defer tc.taskLock.Unlock()
	t, ok := tc.taskMap[id]
	if ok {
		delete(tc.taskMap, id)
		select {
		case t.waitCh <- &msgT[T]{data: msg, err: err}:
			return true
		case <-t.ctx.Done():
			return false
		}
	}
	return false
}

func (tc *TaskCtx[T]) getTask(id any) *task[T] {
	tc.taskLock.Lock()
	defer tc.taskLock.Unlock()
	t := tc.taskMap[id]
	delete(tc.taskMap, id)
	return t
}

func (tc *TaskCtx[T]) setTask(id any, t *task[T]) error {
	tc.taskLock.Lock()
	defer tc.taskLock.Unlock()
	if tc.ctx.Err() != nil {
		return tc.ctx.Err()
	}
	tc.taskMap[id] = t
	return nil
}

func (tc *TaskCtx[T]) delTask(id any) {
	tc.taskLock.Lock()
	defer tc.taskLock.Unlock()
	delete(tc.taskMap, id)
}

type task[T any] struct {
	tctx   *TaskCtx[T]
	ctx    context.Context
	cancel context.CancelFunc

	id any

	initiator Runner

	lock sync.Mutex

	waitCh chan *msgT[T]
}

type Runner func() error

func (t *task[T]) wait(td time.Duration) (msg *T, err error) {
	defer func() {
		if err != nil {
			t.tctx.delTask(t.id)
		}
	}()
	err = t.nowait()
	if err != nil {
		return nil, err
	}
	tk := time.NewTimer(td)
	defer tk.Stop()
	if td <= 0 {
		if !tk.Stop() {
			<-tk.C
		}
	}
	select {
	case <-tk.C:
		err = ErrTaskTimeout
		return nil, err
	case m := <-t.waitCh:
		if m.err != nil {
			return nil, err
		}
		return m.data, nil
	case <-t.ctx.Done():
		return nil, t.ctx.Err()
	}
}

func (t *task[T]) waitCtx(ctx context.Context) (msg *T, err error) {
	defer func() {
		if err != nil {
			t.tctx.delTask(t.id)
		}
	}()
	err = t.nowait()
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case m := <-t.waitCh:
		if m.err != nil {
			return nil, err
		}
		return m.data, nil
	case <-t.ctx.Done():
		return nil, t.ctx.Err()
	}
}

func (t *task[T]) nowait() (err error) {
	if t.ctx.Err() != nil {
		return t.ctx.Err()
	}
	t.lock.Lock()
	defer t.lock.Unlock()

	err = t.tctx.setTask(t.id, t)
	if err != nil {
		return err
	}
	if t.initiator != nil {
		err = t.initiator()
		if err != nil {
			return err
		}
	}

	return err
}

type msgT[T any] struct {
	data *T
	err  error
}
