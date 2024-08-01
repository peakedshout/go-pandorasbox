package xtcp

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func XUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return TCPUpgrader()
}

func TCPUpgrader() xnetutil.Upgrader {
	return &tcpUpgrader{}
}

var TCPConnTmp *net.TCPConn

type tcpUpgrader struct{}

func (t *tcpUpgrader) Upgrade(conn net.Conn) (net.Conn, error) {
	return t.UpgradeContext(context.Background(), conn)
}

func (t *tcpUpgrader) UpgradeContext(ctx context.Context, conn net.Conn) (net.Conn, error) {
	_, err := xnetutil.TypeCheck(conn, TCPConnTmp)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
