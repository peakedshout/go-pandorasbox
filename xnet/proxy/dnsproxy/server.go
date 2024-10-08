package dnsproxy

import (
	"context"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"sync"
	"time"
)

func NewServer(reqCb RespCb, timeout time.Duration) *Server {
	return NewServerContext(context.Background(), reqCb, timeout)
}

func NewServerContext(ctx context.Context, reqCb RespCb, timeout time.Duration) *Server {
	s := &Server{
		timeout: timeout,
		reqCb:   reqCb,
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	return s
}

type RespCb func(ctx context.Context, message *dnsmessage.Message) (*dnsmessage.Message, error)

type Server struct {
	timeout time.Duration
	reqCb   RespCb

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Server) Serve(pconn net.PacketConn) error {
	defer pconn.Close()
	buf := make([]byte, 32*1024)
	sctx := &serverCtx{
		s:    s,
		ponn: pconn,
		uid:  0,
		mux:  sync.Mutex{},
	}
	for {
		n, addr, err := pconn.ReadFrom(buf)
		if err != nil {
			return err
		}
		var message dnsmessage.Message
		err = message.Unpack(buf[:n])
		if err != nil {
			continue
		}
		go sctx.handler(&message, addr)
	}
}

func (s *Server) Close() error {
	s.cancel()
	return s.ctx.Err()
}

type serverCtx struct {
	s    *Server
	ponn net.PacketConn
	uid  uint16
	mux  sync.Mutex
}

func (sctx *serverCtx) newId() uint16 {
	sctx.mux.Lock()
	defer sctx.mux.Unlock()
	sctx.uid++
	return sctx.uid
}

func (sctx *serverCtx) handler(message *dnsmessage.Message, addr net.Addr) {
	// server not need response
	if message.Header.Response {
		return
	} else {
		rid := message.ID
		nid := sctx.newId()
		dt := newDnsTask(nid, func(message *dnsmessage.Message) {
			message.ID = rid
			pack, _ := message.Pack()
			_, _ = sctx.ponn.WriteTo(pack, addr)
		})
		ctx := sctx.s.ctx
		var cancelFunc context.CancelFunc
		if sctx.s.timeout != 0 {
			ctx, cancelFunc = context.WithTimeout(ctx, sctx.s.timeout)
			defer cancelFunc()
		} else {
			ctx, cancelFunc = context.WithCancel(ctx)
			defer cancelFunc()
		}

		message.ID = nid
		if sctx.s.reqCb != nil {
			resp, err := sctx.s.reqCb(ctx, message)
			if err != nil {
				return
			}
			dt.Resp(resp)
		}
	}
}

func newDnsTask(uid uint16, resp func(message *dnsmessage.Message)) *dnsTask {
	return &dnsTask{
		respFn: resp,
		uid:    &uid,
	}
}

type dnsTask struct {
	respFn    func(message *dnsmessage.Message)
	timeoutFn func()
	uid       *uint16
}

func (d *dnsTask) Resp(message *dnsmessage.Message) {
	if d.respFn != nil {
		d.respFn(message)
	}
}
