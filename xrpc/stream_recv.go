package xrpc

import (
	"context"
	"errors"
	"github.com/peakedshout/go-pandorasbox/xmsg"
	"sync"
)

type RecvStreamHandler func(ctx RecvStream) (any, error)

type recvServerStream struct {
	stream *serverStream
}

func (rsc *recvServerStream) Recv(out any) error {
	return rsc.stream.Recv(out)
}

func (rsc *recvServerStream) Context() context.Context {
	return rsc.stream.ctx
}

func (rsc *recvServerStream) Close() error {
	return rsc.stream.Close()
}

type sendClientStream struct {
	cs   *clientStream
	mux  sync.Mutex
	xMsg *xmsg.XMsg
}

func (scs *sendClientStream) Send(data any) error {
	return scs.cs.Send(data)
}

func (scs *sendClientStream) Close() error {
	return scs.cs.Close()
}

func (scs *sendClientStream) Context() context.Context {
	return scs.cs.Context()
}

func (scs *sendClientStream) run() {
	scs.mux.Lock()
	defer scs.mux.Unlock()
	defer scs.cs.Close()
	select {
	case <-scs.cs.ctx.Done():
	case xMsg, ok := <-scs.cs.read:
		if !ok {
			return
		}
		scs.xMsg = xMsg
	}
}

func (scs *sendClientStream) Bind(out any) error {
	scs.mux.Lock()
	defer scs.mux.Unlock()
	if scs.xMsg == nil {
		return errors.New("nil data")
	}
	return scs.xMsg.Unmarshal(out)
}
