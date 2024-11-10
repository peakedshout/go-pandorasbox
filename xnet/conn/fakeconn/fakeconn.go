package fakeconn

import (
	"context"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

func NewFakeConn(ctx context.Context, r, l net.Addr, rdata chan []byte, wdata chan *AddrData) (net.Conn, context.Context) {
	xctx, cl := context.WithCancelCause(ctx)
	xc := &fakeConn{
		raddr:  r,
		laddr:  l,
		bctx:   context.Background(),
		ctx:    xctx,
		cl:     cl,
		rdata:  rdata,
		rsleep: make(chan struct{}, 1),
		wdata:  wdata,
		wsleep: make(chan struct{}, 1),
	}
	return xc, xctx
}

type fakeConn struct {
	raddr, laddr net.Addr
	mux          sync.Mutex
	bctx         context.Context
	ctx          context.Context
	cl           context.CancelCauseFunc
	rdata        chan []byte
	rlast        []byte
	rctx         context.Context
	rcl          context.CancelFunc
	rmux         sync.Mutex
	rsleep       chan struct{}
	wdata        chan *AddrData
	wctx         context.Context
	wcl          context.CancelFunc
	wmux         sync.Mutex
	wsleep       chan struct{}
}

func (x *fakeConn) Read(b []byte) (n int, err error) {
	x.rmux.Lock()
	defer x.rmux.Unlock()
	for {
		if x.ctx.Err() != nil {
			return 0, context.Cause(x.ctx)
		}
		x.mux.Lock()
		ctx := x.rctx
		x.mux.Unlock()
		if ctx == nil {
			ctx = x.bctx
		}
		if ctx.Err() != nil {
			return 0, os.ErrDeadlineExceeded
		}
		if len(x.rlast) != 0 {
			n = copy(b, x.rlast)
			x.rlast = x.rlast[n:]
			return n, nil
		}
		select {
		case x.rlast = <-x.rdata:
			n = copy(b, x.rlast)
			x.rlast = x.rlast[n:]
			return n, nil
		case <-x.rsleep:
			continue
		case <-ctx.Done():
			return 0, os.ErrDeadlineExceeded
		case <-x.ctx.Done():
			return 0, context.Cause(x.ctx)
		}
	}
}

func (x *fakeConn) Write(b []byte) (n int, err error) {
	x.wmux.Lock()
	defer x.wmux.Unlock()
	for {
		if x.ctx.Err() != nil {
			return 0, context.Cause(x.ctx)
		}
		x.mux.Lock()
		ctx := x.wctx
		x.mux.Unlock()
		if ctx == nil {
			ctx = x.bctx
		}
		if ctx.Err() != nil {
			return 0, os.ErrDeadlineExceeded
		}
		select {
		case x.wdata <- &AddrData{
			Addr: x.raddr,
			Data: b,
		}:
			return len(b), nil
		case <-x.wsleep:
			continue
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-x.ctx.Done():
			return 0, context.Cause(x.ctx)
		}
	}
}

func (x *fakeConn) Close() error {
	if x.ctx.Err() != nil {
		return context.Cause(x.ctx)
	}
	x.cl(io.EOF)
	return nil
}

func (x *fakeConn) LocalAddr() net.Addr {
	return x.laddr
}

func (x *fakeConn) RemoteAddr() net.Addr {
	return x.raddr
}

func (x *fakeConn) SetDeadline(t time.Time) error {
	x.mux.Lock()
	defer x.mux.Unlock()
	_ = x.setReadDeadline(t)
	_ = x.setWriteDeadline(t)
	return nil
}

func (x *fakeConn) SetReadDeadline(t time.Time) error {
	x.mux.Lock()
	defer x.mux.Unlock()
	_ = x.setReadDeadline(t)
	return nil
}

func (x *fakeConn) SetWriteDeadline(t time.Time) error {
	x.mux.Lock()
	defer x.mux.Unlock()
	_ = x.setWriteDeadline(t)
	return nil
}

func (x *fakeConn) setReadDeadline(t time.Time) error {
	select {
	case x.rsleep <- struct{}{}:
	default:
	}
	if x.rcl != nil {
		x.rcl()
	}
	if t.IsZero() {
		x.rctx, x.rcl = nil, nil
	} else {
		x.rctx, x.rcl = context.WithDeadline(context.Background(), t)
	}
	return nil
}

func (x *fakeConn) setWriteDeadline(t time.Time) error {
	select {
	case x.wsleep <- struct{}{}:
	default:
	}
	if x.wcl != nil {
		x.wcl()
	}
	if t.IsZero() {
		x.wctx, x.wcl = nil, nil
	} else {
		x.wctx, x.wcl = context.WithDeadline(context.Background(), t)
	}
	return nil
}

type AddrData struct {
	Addr net.Addr
	Data []byte
}
