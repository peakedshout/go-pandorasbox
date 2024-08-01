package xnetutil

import (
	"context"
	"net"
)

type Dialer interface {
	Dial(network string, addr string) (net.Conn, error)
	DialContext(ctx context.Context, network string, addr string) (net.Conn, error)
}

func NewCallBackDialer(fn func(ctx context.Context, network string, addr string) (net.Conn, error)) Dialer {
	return &cbDialer{cbFn: fn}
}

type cbDialer struct {
	cbFn func(ctx context.Context, network string, addr string) (net.Conn, error)
}

func (cd *cbDialer) Dial(network string, addr string) (net.Conn, error) {
	return cd.DialContext(context.Background(), network, addr)
}

func (cd *cbDialer) DialContext(ctx context.Context, network string, addr string) (net.Conn, error) {
	return cd.cbFn(ctx, network, addr)
}
