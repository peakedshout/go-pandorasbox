package xrpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/peakedshout/go-pandorasbox/xmsg"
	"sync"
	"sync/atomic"
)

type ReverseRpcHandler func(ctx ReverseRpc) error

type serverReverseRpcContext struct {
	xMsg   *xmsg.XMsg
	stream *serverStream
	uid    uint32
	rpcMux sync.Mutex
	rpcMap map[uint32]chan *rRpcMsg
}

func (rrc *serverReverseRpcContext) Bind(out any) error {
	return rrc.xMsg.Unmarshal(out)
}

func (rrc *serverReverseRpcContext) Context() context.Context {
	return rrc.stream.Context()
}

func (rrc *serverReverseRpcContext) Rpc(ctx context.Context, header string, send, recv any) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	data, err := rrc.marshalData(header, send)
	if err != nil {
		return err
	}
	err = rrc.stream.Send(data)
	if err != nil {
		return err
	}
	ch := make(chan *rRpcMsg, 1)
	rrc.rpcMux.Lock()
	rrc.rpcMap[data.Id] = ch
	rrc.rpcMux.Unlock()
	defer func() {
		rrc.rpcMux.Lock()
		delete(rrc.rpcMap, data.Id)
		rrc.rpcMux.Unlock()
	}()
	select {
	case msg := <-ch:
		if recv == nil {
			return nil
		}
		return msg.unmarshal(recv)
	case <-ctx.Done():
		return ctx.Err()
	case <-rrc.stream.Context().Done():
		return rrc.stream.Context().Err()
	}
}

func (rrc *serverReverseRpcContext) handleXMsg() {
	defer rrc.stream.close(ErrRRpcClosed)
	for {
		msg := new(rRpcMsg)
		err := rrc.stream.Recv(msg)
		if err != nil {
			return
		}
		rrc.rpcMux.Lock()
		ch, ok := rrc.rpcMap[msg.Id]
		if ok {
			ch <- msg
			delete(rrc.rpcMap, msg.Id)
		}
		rrc.rpcMux.Unlock()
	}
}

func (rrc *serverReverseRpcContext) marshalData(header string, send any) (*rRpcMsg, error) {
	var id uint32
	for ; id == 0; id = atomic.AddUint32(&rrc.uid, 1) {
	}
	err, ok := send.(error)
	if ok {
		return &rRpcMsg{
			Header: header,
			Id:     id,
			Type:   rMsgTypeErr,
			Data:   []byte(err.Error()),
		}, nil
	}
	bytes, err := json.Marshal(send)
	if err != nil {
		return nil, err
	}
	return &rRpcMsg{
		Header: header,
		Id:     id,
		Type:   rMsgTypeMsg,
		Data:   bytes,
	}, nil
}

type rRpcMsg struct {
	Header string
	Id     uint32
	Type   uint8
	Data   []byte
}

const (
	rMsgTypeMsg = iota
	rMsgTypeErr
)

func (r *rRpcMsg) unmarshal(out any) error {
	if len(r.Data) == 0 {
		return errors.New("nil data")
	}
	switch r.Type {
	case rMsgTypeMsg:
		return json.Unmarshal(r.Data, out)
	case rMsgTypeErr:
		var str string
		err := json.Unmarshal(r.Data, &str)
		if err != nil {
			return err
		}
		return errors.New(str)
	default:
		return errors.New("invalid rRpcMsg")
	}
}

type ClientReverseRpcHandler func(ClientReverseRpcContext) (any, error)

type ClientReverseRpcContext interface {
	Bind(out any) error
	Context() context.Context
}

type clientReverseRpcContext struct {
	ctx context.Context
	msg *rRpcMsg
}

func (c *clientReverseRpcContext) Bind(out any) error {
	return c.msg.unmarshal(out)
}

func (c *clientReverseRpcContext) Context() context.Context {
	return c.ctx
}
