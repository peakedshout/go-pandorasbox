package xwebsocket

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/xnet/conn/websocketconn"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
)

func XUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return WebsocketUpgrader(cfg, isClient)
}

func WebsocketUpgrader(cfg *tls.Config, isClient bool) xnetutil.Upgrader {
	return &websocketUpgrader{
		cfg:      cfg,
		isClient: isClient,
	}
}

var WebsocketConnTmp *websocketconn.WebsocketConn

type websocketUpgrader struct {
	cfg      *tls.Config
	isClient bool
}

func (w *websocketUpgrader) Upgrade(conn net.Conn) (net.Conn, error) {
	return w.UpgradeContext(context.Background(), conn)
}

func (w *websocketUpgrader) UpgradeContext(ctx context.Context, conn net.Conn) (net.Conn, error) {
	if w.isClient {
		return websocketconn.Client(conn, w.cfg), nil
	} else {
		return websocketconn.Server(conn, w.cfg), nil
	}
}
