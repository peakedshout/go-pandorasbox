package xtls

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func XUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return TLSUpgrader(cfg, isClient)
}

func TLSUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return &tlsUpgrader{
		cfg:      cfg,
		isClient: isClient,
	}
}

var TLSConnTmp *tls.Conn

type tlsUpgrader struct {
	cfg      *tls.Config
	isClient bool
}

func (t *tlsUpgrader) Upgrade(conn net.Conn) (net.Conn, error) {
	return t.UpgradeContext(context.Background(), conn)
}

func (t *tlsUpgrader) UpgradeContext(ctx context.Context, conn net.Conn) (net.Conn, error) {
	if t.isClient {
		return tls.Client(conn, t.cfg), nil
	} else {
		return tls.Server(conn, t.cfg), nil
	}
}
