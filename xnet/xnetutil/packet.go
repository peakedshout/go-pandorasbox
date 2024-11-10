package xnetutil

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/xnet/conn/fakeconn"
	"net"
	"sync"
	"time"
)

func NewXPacketListenCfg(gc, timeout time.Duration) *XPacketListenCfg {
	return &XPacketListenCfg{GCDuration: gc, TimeoutDuration: timeout}
}

type XPacketListenCfg struct {
	GCDuration      time.Duration
	TimeoutDuration time.Duration
}

func (p *XPacketListenCfg) Listen(network string, address string) (net.Listener, error) {
	return p.ListenContext(context.Background(), network, address)
}

func (p *XPacketListenCfg) ListenContext(ctx context.Context, network string, address string) (net.Listener, error) {
	pconn, err := NewDefaultListenerConfig(nil).ListenPacketContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	listener := newPacketListener(ctx, pconn, p.GCDuration, p.TimeoutDuration)
	return listener, nil
}

func newPacketListener(ctx context.Context, conn net.PacketConn, gc, timeout time.Duration) net.Listener {
	ln := &xPacketListener{
		conn:            conn,
		wchan:           make(chan *fakeconn.AddrData),
		ch:              make(chan net.Conn),
		connMap:         make(map[string]*connInfo),
		gcDuration:      gc,
		timeoutDuration: timeout,
	}
	ln.ctx, ln.cl = context.WithCancel(ctx)
	go ln.gcHandle()
	go ln.writeHandle()
	go ln.readHandle()
	return ln
}

type xPacketListener struct {
	conn            net.PacketConn
	mux             sync.Mutex
	wchan           chan *fakeconn.AddrData
	ch              chan net.Conn
	ctx             context.Context
	cl              context.CancelFunc
	connMap         map[string]*connInfo
	gcDuration      time.Duration
	timeoutDuration time.Duration
}

func (x *xPacketListener) Accept() (net.Conn, error) {
	select {
	case <-x.ctx.Done():
		return nil, x.ctx.Err()
	case conn := <-x.ch:
		return conn, nil
	}
}

func (x *xPacketListener) Close() error {
	x.cl()
	return x.conn.Close()
}

func (x *xPacketListener) Addr() net.Addr {
	return x.conn.LocalAddr()
}

func (x *xPacketListener) gcHandle() {
	if x.gcDuration <= 0 {
		x.gcDuration = 30 * time.Second
	}
	if x.timeoutDuration <= 0 {
		x.timeoutDuration = 30 * time.Second
	}
	ticker := time.NewTicker(x.gcDuration)
	defer ticker.Stop()
	for {
		select {
		case <-x.ctx.Done():
			return
		case t := <-ticker.C:
			t = t.Add(-x.timeoutDuration)
			x.mux.Lock()
			var dl []string
			for k, info := range x.connMap {
				if info.lastT.Before(t) {
					dl = append(dl, k)
				}
			}
			for _, k := range dl {
				info := x.connMap[k]
				_ = info.Close()
				delete(x.connMap, k)
			}
			x.mux.Unlock()
		}
	}

}

func (x *xPacketListener) writeHandle() {
	defer x.Close()
	for {
		select {
		case <-x.ctx.Done():
			return
		case data := <-x.wchan:
			str := data.Addr.String()
			x.mux.Lock()
			c := x.connMap[str]
			if c != nil {
				c.lastT = time.Now()
			}
			x.mux.Unlock()
			_, err := x.conn.WriteTo(data.Data, data.Addr)
			if err != nil {
				_ = c.Close()
				continue
			}
		}
	}
}

func (x *xPacketListener) readHandle() {
	defer x.Close()
	buf := make([]byte, 32*1024)
	for {
		n, addr, err := x.conn.ReadFrom(buf)
		if err != nil {
			return
		}
		x.readData(buf[:n], addr)
	}
}

func (x *xPacketListener) readData(b []byte, addr net.Addr) {
	x.mux.Lock()
	defer x.mux.Unlock()
	str := addr.String()
	c, ok := x.connMap[str]
	if !ok {
		c = x.newConn(addr)
	}
	if c != nil {
		c.lastT = time.Now()
		c.read(b)
	}
}

func (x *xPacketListener) newConn(addr net.Addr) *connInfo {
	cinfo := &connInfo{
		ch:  make(chan []byte, 1),
		ctx: nil,
	}
	cinfo.Conn, cinfo.ctx = fakeconn.NewFakeConn(x.ctx, addr, x.conn.LocalAddr(), cinfo.ch, x.wchan)
	select {
	case <-x.ctx.Done():
		return nil
	case x.ch <- cinfo.Conn:
	}
	x.connMap[addr.String()] = cinfo
	return cinfo
}

type connInfo struct {
	net.Conn
	ch    chan []byte
	ctx   context.Context
	lastT time.Time
}

func (c *connInfo) read(b []byte) {
	select {
	case c.ch <- b:
	case <-c.ctx.Done():
	}
}
