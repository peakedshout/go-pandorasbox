package ctxtool

import (
	"context"
	"reflect"
)

func ContextsWithCancel(ctx context.Context, ctxs ...context.Context) (context.Context, context.CancelFunc) {
	rCtx, cancel := context.WithCancelCause(ctx)
	nCtx := &contexts{
		Context: rCtx,
		ctxs:    checkCtx(ctxs...),
		cancel:  cancel,
	}
	nCtx.run()
	return nCtx, func() { cancel(context.Canceled) }
}

func ContextsWithCancelCause(ctx context.Context, ctxs ...context.Context) (context.Context, context.CancelCauseFunc) {
	rCtx, cancel := context.WithCancelCause(ctx)
	nCtx := &contexts{
		Context: rCtx,
		ctxs:    checkCtx(ctxs...),
		cancel:  cancel,
	}
	nCtx.run()
	return nCtx, cancel
}

func checkCtx(ctxs ...context.Context) []context.Context {
	list := make([]context.Context, 0, len(ctxs))
	for _, ctx := range ctxs {
		if ctx == context.Background() || ctx == context.TODO() {
			continue
		}
		list = append(list, ctx)
	}
	return list
}

type contexts struct {
	context.Context
	cancel context.CancelCauseFunc
	ctxs   []context.Context
}

func (c *contexts) Err() error {
	return context.Cause(c.Context)
}

func (c *contexts) run() {
	if len(c.ctxs) > 0 {
		cases := make([]reflect.SelectCase, 0, len(c.ctxs)+1)
		cases = append(cases, toCases(c.Context.Done()))
		for _, one := range c.ctxs {
			cases = append(cases, toCases(one.Done()))
		}
		go func() {
			chosen, _, _ := reflect.Select(cases)
			if chosen != 0 {
				c.cancel(c.ctxs[chosen-1].Err())
			}
		}()
	}
}

func toCases(ch <-chan struct{}) reflect.SelectCase {
	return reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
}

func WaitContexts(ctxs ...context.Context) error {
	if len(ctxs) == 1 {
		return Wait(ctxs[0])
	} else if len(ctxs) > 0 {
		cases := make([]reflect.SelectCase, 0, len(ctxs))
		for _, one := range ctxs {
			cases = append(cases, toCases(one.Done()))
		}
		chosen, _, _ := reflect.Select(cases)
		return ctxs[chosen].Err()
	}
	return nil
}

func WaitContextsFunc(fn func(), ctxs ...context.Context) {
	_ = WaitContexts(ctxs...)
	fn()
}
