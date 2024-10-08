package xdummy

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func NewDummyListenDialer(ctx context.Context) (net.Listener, xnetutil.Dialer) {
	ch := make(chan net.Conn)
	dialer := NewDummyDialer(ctx, ch)
	listener := NewDummyListener(ctx, ch)
	return listener, dialer
}
