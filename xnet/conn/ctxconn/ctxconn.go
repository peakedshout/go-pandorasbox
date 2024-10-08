package ctxconn

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"net"
)

func NewCtxConn(ctx context.Context, conn net.Conn) context.Context {
	nCtx, _ := monitorConn(ctx, conn)
	return nCtx
}

func NewCtxConnWithCancel(ctx context.Context, conn net.Conn) (context.Context, context.CancelFunc) {
	return monitorConn(ctx, conn)
}

func monitorConn(ctx context.Context, conn net.Conn) (context.Context, context.CancelFunc) {
	return ctxtool.NewRcContextWithCancel(ctx, conn)
}
