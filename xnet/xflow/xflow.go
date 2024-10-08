package xflow

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/xnet/conn/flowconn"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func XUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return FlowUpgrader()
}

func FlowUpgrader() xnetutil.Upgrader {
	return &flowUpgrader{}
}

var FlowConnTmp *flowconn.Conn

type flowUpgrader struct{}

func (f *flowUpgrader) Upgrade(conn net.Conn) (net.Conn, error) {
	return f.UpgradeContext(context.Background(), conn)
}

func (f *flowUpgrader) UpgradeContext(ctx context.Context, conn net.Conn) (net.Conn, error) {
	return flowconn.NewConn(conn), nil
}
