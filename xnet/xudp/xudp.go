package xudp

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func XUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return UDPUpgrader()
}

func UDPUpgrader() xnetutil.Upgrader {
	return &udpUpgrager{}
}

type udpUpgrager struct{}

var UDPConnTmp *net.UDPConn

func (u *udpUpgrager) Upgrade(conn net.Conn) (net.Conn, error) {
	return u.UpgradeContext(context.Background(), conn)
}

func (u *udpUpgrager) UpgradeContext(ctx context.Context, conn net.Conn) (net.Conn, error) {
	_, err := xnetutil.TypeCheck(conn, UDPConnTmp)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
