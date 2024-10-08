package xdummy

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func NewDummyDialer(ctx context.Context, ch chan net.Conn) xnetutil.Dialer {
	return &dummyDialer{ctx: ctx, ch: ch}
}

type dummyDialer struct {
	ctx context.Context
	ch  chan net.Conn
}

func (d *dummyDialer) Dial(network string, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}

func (d *dummyDialer) DialContext(ctx context.Context, network string, addr string) (net.Conn, error) {
	c1, c2 := NewDummyConn()
	select {
	case d.ch <- c2:
		return c1, nil
	case <-ctx.Done():
		_ = c1.Close()
		_ = c2.Close()
		return nil, ctx.Err()
	case <-d.ctx.Done():
		_ = c1.Close()
		_ = c2.Close()
		return nil, d.ctx.Err()
	}
}
