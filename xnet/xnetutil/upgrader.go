package xnetutil

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/tool/gpool"
	"net"
	"sync"
)

// Upgrader
//
//	In fact, it should be a cold update, such as handshake operations, which should not be processed here,
//	and handshakes are not provided internally. For the sake of simplicity,
//	the internal implementation should just build the structure of conn, rather than handle the complex logic.
type Upgrader interface {
	Upgrade(conn net.Conn) (net.Conn, error)
	UpgradeContext(ctx context.Context, conn net.Conn) (net.Conn, error)
}

func EmptyUpgrader() Upgrader {
	return &emptyUpgrader{}
}

type emptyUpgrader struct{}

func (eu *emptyUpgrader) Upgrade(conn net.Conn) (net.Conn, error) {
	return conn, nil
}

func (eu *emptyUpgrader) UpgradeContext(ctx context.Context, conn net.Conn) (net.Conn, error) {
	return conn, nil
}

type warpUpgrader struct {
	upgraderList []Upgrader
}

func NewWarpUpgrader(uList ...Upgrader) Upgrader {
	return &warpUpgrader{upgraderList: uList}
}

func (wu *warpUpgrader) Upgrade(conn net.Conn) (uconn net.Conn, err error) {
	return wu.UpgradeContext(context.Background(), conn)
}

func (wu *warpUpgrader) UpgradeContext(ctx context.Context, conn net.Conn) (uconn net.Conn, err error) {
	for _, upgrader := range wu.upgraderList {
		conn, err = upgrader.UpgradeContext(ctx, conn)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func (wu *warpUpgrader) Add(uList ...Upgrader) {
	wu.upgraderList = append(wu.upgraderList, uList...)
}

type upgraderListener struct {
	upgrader Upgrader
	ln       net.Listener

	ctx    context.Context
	cancel context.CancelFunc
	g      *gpool.GPool
	ch     chan net.Conn
	once   sync.Once
}

func NewUpgraderListener(ln net.Listener, upgrader Upgrader, g int, c int) net.Listener {
	ul := &upgraderListener{
		upgrader: upgrader,
		ln:       ln,
	}
	if c < 0 {
		c = 0
	}
	ul.ch = make(chan net.Conn, c)
	ul.ctx, ul.cancel = context.WithCancel(context.Background())
	ul.g = gpool.NewGPool(ul.ctx, g)
	return ul
}

func (ul *upgraderListener) Accept() (net.Conn, error) {
	ul.once.Do(func() {
		go ul.run()
	})
	select {
	case conn := <-ul.ch:
		return conn, nil
	case <-ul.ctx.Done():
		ul.Close()
		return nil, ul.ctx.Err()
	}
}

func (ul *upgraderListener) Close() error {
	ul.cancel()
	return ul.ln.Close()
}

func (ul *upgraderListener) Addr() net.Addr {
	return ul.ln.Addr()
}

func (ul *upgraderListener) run() {
	defer ul.Close()
	for {
		conn, err := ul.ln.Accept()
		if err != nil {
			return
		}
		if !ul.g.Do(func() {
			upgrade, err := ul.upgrader.Upgrade(conn)
			if err != nil {
				return
			}
			select {
			case ul.ch <- upgrade:
			case <-ul.ctx.Done():
			}
		}) {
			return
		}
	}
}

type upgraderDialer struct {
	upgrader Upgrader
	dr       Dialer
}

func NewUpgraderDialer(upgrader Upgrader, dr Dialer) Dialer {
	return &upgraderDialer{
		upgrader: upgrader,
		dr:       dr,
	}
}

func (ud *upgraderDialer) Dial(network string, add string) (net.Conn, error) {
	return ud.DialContext(context.Background(), network, add)
}

func (ud *upgraderDialer) DialContext(ctx context.Context, network string, add string) (net.Conn, error) {
	conn, err := ud.dr.DialContext(ctx, network, add)
	if err != nil {
		return nil, err
	}
	return ud.upgrader.UpgradeContext(ctx, conn)
}
