package xfakehttp

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/xnet/conn/fakehttpconn"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func XUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return FakeHttpUpgrader(cfg, isClient)
}

func FakeHttpUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return &fakeHttpUpgrader{
		cfg:      cfg,
		isClient: isClient,
	}
}

var FakeHttpConnTmp *fakehttpconn.FakeHttpConn

type fakeHttpUpgrader struct {
	cfg      *tls.Config
	isClient bool
}

func (f *fakeHttpUpgrader) Upgrade(conn net.Conn) (net.Conn, error) {
	return f.UpgradeContext(context.Background(), conn)
}

func (f *fakeHttpUpgrader) UpgradeContext(ctx context.Context, conn net.Conn) (net.Conn, error) {
	if f.isClient {
		return tls.Client(conn, f.cfg), nil
	} else {
		return tls.Server(conn, f.cfg), nil
	}
}
