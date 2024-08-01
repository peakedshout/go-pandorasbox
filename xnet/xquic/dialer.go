package xquic

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/xnet/conn/quicconn"
	"github.com/quic-go/quic-go"
	"net"
	"time"
)

var _defaultDialer = NewDialer(nil, _defaultQuicTlsConfg)

func Dial(network string, addr string) (conn net.Conn, err error) {
	return _defaultDialer.Dial(network, addr)
}

func DialContext(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
	return _defaultDialer.DialContext(ctx, network, addr)
}

func NewDialer(c net.PacketConn, tlsCfg *tls.Config) *QuicDialer {
	if tlsCfg == nil {
		tlsCfg = _defaultQuicTlsConfg
	}
	return &QuicDialer{
		c:      c,
		tlsCfg: tlsCfg,
	}
}

const XQuicPreHandShake = "XQuicPreHandShake"

type QuicDialer struct {
	c      net.PacketConn
	tlsCfg *tls.Config
}

func (qd *QuicDialer) Dial(network string, addr string) (conn net.Conn, err error) {
	return qd.DialContext(context.Background(), network, addr)
}

func (qd *QuicDialer) DialContext(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
	var connection quic.Connection
	if qd.c == nil {
		connection, err = quic.DialAddr(ctx, addr, qd.tlsCfg, nil)
		if err != nil {
			return
		}
	} else {
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return nil, err
		}
		connection, err = quic.Dial(ctx, qd.c, udpAddr, qd.tlsCfg, nil)
		if err != nil {
			return nil, err
		}
	}
	conn = quicconn.NewConn(true, connection)
	value := ctx.Value(XQuicPreHandShake)
	if value != nil {
		err = conn.SetWriteDeadline(time.Time{})
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func NewQuicTransportDialer(tr *quic.Transport, tlsCfg *tls.Config) *QuicTransportDialer {
	if tlsCfg == nil {
		tlsCfg = _defaultQuicTlsConfg
	}
	return &QuicTransportDialer{
		tr:     tr,
		tlsCfg: tlsCfg,
	}
}

type QuicTransportDialer struct {
	tr     *quic.Transport
	tlsCfg *tls.Config
}

func (qtdr *QuicTransportDialer) Dial(network string, addr string) (net.Conn, error) {
	return qtdr.DialContext(context.Background(), network, addr)
}

func (qtdr *QuicTransportDialer) DialContext(ctx context.Context, network string, addr string) (net.Conn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	connection, err := qtdr.tr.Dial(ctx, udpAddr, qtdr.tlsCfg, nil)
	if err != nil {
		return nil, err
	}
	conn := quicconn.NewConn(true, connection)
	value := ctx.Value(XQuicPreHandShake)
	if value != nil {
		err = conn.SetWriteDeadline(time.Time{})
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func (qtdr *QuicTransportDialer) Close() error {
	return qtdr.tr.Close()
}

func (qtdr *QuicTransportDialer) Transport() *quic.Transport {
	return qtdr.tr
}
