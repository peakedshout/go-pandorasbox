package xquic

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/xnet/conn/quicconn"
	"github.com/quic-go/quic-go"
	"net"
)

var _defaultListenConfig = NewQuicListenConfig(nil, _defaultQuicTlsConfg)

func Listen(network, addr string) (ln net.Listener, err error) {
	return _defaultListenConfig.Listen(network, addr)
}

func ListenContext(ctx context.Context, network, addr string) (ln net.Listener, err error) {
	return _defaultListenConfig.ListenContext(ctx, network, addr)
}

func NewQuicListenConfig(c net.PacketConn, tlsCfg *tls.Config) *QuicListenConfig {
	if tlsCfg == nil {
		tlsCfg = _defaultQuicTlsConfg
	}
	return &QuicListenConfig{
		c:      c,
		tlsCfg: tlsCfg,
	}
}

func NewQuicListenConfigWithTransport(tr *quic.Transport, tlsCfg *tls.Config) *QuicListenConfig {
	if tlsCfg == nil {
		tlsCfg = _defaultQuicTlsConfg
	}
	return &QuicListenConfig{
		tr:     tr,
		tlsCfg: tlsCfg,
	}
}

type QuicListenConfig struct {
	tr     *quic.Transport
	c      net.PacketConn
	tlsCfg *tls.Config
}

// Listen if exist c will be not use address and network
func (qc *QuicListenConfig) Listen(network string, address string) (net.Listener, error) {
	return qc.ListenContext(context.Background(), network, address)
}

// ListenContext if exist c will be not use address and network
func (qc *QuicListenConfig) ListenContext(ctx context.Context, network string, addr string) (ln net.Listener, err error) {
	var listen *quic.Listener
	if qc.c != nil {
		listen, err = quic.Listen(qc.c, qc.tlsCfg, nil)
		if err != nil {
			return nil, err
		}
	} else if qc.tr != nil {
		listen, err = qc.tr.Listen(qc.tlsCfg, nil)
		if err != nil {
			return nil, err
		}
	} else {
		listen, err = quic.ListenAddr(addr, qc.tlsCfg, nil)
		if err != nil {
			return nil, err
		}
	}
	return &quicListener{ln: listen, ctx: ctx}, nil
}

type quicListener struct {
	ctx context.Context
	ln  *quic.Listener
}

func (ql *quicListener) Accept() (net.Conn, error) {
	connection, err := ql.ln.Accept(ql.ctx)
	if err != nil {
		return nil, err
	}
	return quicconn.NewConn(false, connection), nil
}

func (ql *quicListener) Close() error {
	return ql.ln.Close()
}

func (ql *quicListener) Addr() net.Addr {
	return ql.ln.Addr()
}
