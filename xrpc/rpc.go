package xrpc

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/xmsg"
)

type rpcContext struct {
	ctx  context.Context
	xMsg *xmsg.XMsg
}

func (rc *rpcContext) Bind(out any) error {
	return rc.xMsg.Unmarshal(out)
}

func (rc *rpcContext) Context() context.Context {
	return rc.ctx
}

type RpcHandler func(ctx Rpc) (any, error)
