package xquic

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/xnet/conn/quicconn"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func XUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return QuicUpgrader()
}

func QuicUpgrader() xnetutil.Upgrader {
	return &quicUpgrader{}
}

var _defaultQuicTlsConfg = pcrypto.MustNewDefaultTlsConfig()

var QuicConnTmp *quicconn.QuicConn

type quicUpgrader struct{}

func (q *quicUpgrader) Upgrade(conn net.Conn) (net.Conn, error) {
	return q.UpgradeContext(context.Background(), conn)
}

func (q *quicUpgrader) UpgradeContext(ctx context.Context, conn net.Conn) (net.Conn, error) {
	_, err := xnetutil.TypeCheck(conn, QuicConnTmp)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
