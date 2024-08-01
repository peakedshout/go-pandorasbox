package xdummy

import (
	"context"
	"errors"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func NewDummyListener(ctx context.Context, ch chan net.Conn) net.Listener {
	cctx, cl := context.WithCancelCause(ctx)
	return &dummyListener{
		ctx: cctx,
		cl:  cl,
		ch:  ch,
	}
}

type dummyListener struct {
	ctx context.Context
	cl  context.CancelCauseFunc
	ch  chan net.Conn
}

func (l *dummyListener) Accept() (net.Conn, error) {
	select {
	case _conn := <-l.ch:
		return _conn, nil
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	}
}

func (l *dummyListener) Close() error {
	l.cl(errors.New("closed"))
	return nil
}

func (l *dummyListener) Addr() net.Addr {
	return xnetutil.NewNetAddr("dummy", "dummy")
}
