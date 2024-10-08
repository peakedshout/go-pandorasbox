package xrpc

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/xmsg"
)

type SendStreamHandler func(ctx SendStream) error

type sendStreamContext struct {
	xMsg   *xmsg.XMsg
	stream *serverStream
}

func (sc *sendStreamContext) Send(data any) error {
	return sc.stream.Send(data)
}

func (sc *sendStreamContext) Context() context.Context {
	return sc.stream.ctx
}

func (sc *sendStreamContext) Bind(out any) error {
	return sc.xMsg.Unmarshal(out)
}

func (sc *sendStreamContext) Close() error {
	return sc.stream.Close()
}

type recvClientStream struct {
	cs *clientStream
}

func (rcs *recvClientStream) Recv(out any) error {
	return rcs.cs.Recv(out)
}

func (rcs *recvClientStream) Close() error {
	return rcs.cs.Close()
}

func (rcs *recvClientStream) Context() context.Context {
	return rcs.cs.Context()
}
