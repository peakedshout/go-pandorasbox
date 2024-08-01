package xspeed

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/xnet/conn/speedconn"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func XUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return SpeedPUpgrader()
}

func SpeedPUpgrader() xnetutil.Upgrader {
	return &speedUpgrader{}
}

var SpeedConnTmp *speedconn.Conn

type speedUpgrader struct{}

func (s *speedUpgrader) Upgrade(conn net.Conn) (net.Conn, error) {
	return s.UpgradeContext(context.Background(), conn)
}

func (s *speedUpgrader) UpgradeContext(ctx context.Context, conn net.Conn) (net.Conn, error) {
	return speedconn.NewConn(conn), nil
}
