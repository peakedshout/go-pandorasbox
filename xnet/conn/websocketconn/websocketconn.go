package websocketconn

import (
	"bufio"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/xnet/xtls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type WebsocketConn struct {
	isClient bool
	conn     net.Conn
	tlsCfg   *tls.Config
	reader   *bufio.Reader
	writer   *bufio.Writer
	laddr    net.Addr
	raddr    net.Addr
	once     sync.Once
	mux      sync.Mutex

	accept []byte

	version       int
	location      *url.URL
	origin        *url.URL
	protocol      []string
	header        http.Header
	handshakeData map[string]string

	rio sync.Mutex
	wio sync.Mutex
	frameReaderFactory
	frameReader
	frameWriterFactory
	frameHandler
	payloadType        byte
	defaultCloseStatus int
	MaxPayloadBytes    int
}

func (wc *WebsocketConn) Read(b []byte) (n int, err error) {
	err = wc.HandShakeHandle()
	if err != nil {
		return 0, err
	}
	wc.rio.Lock()
	defer wc.rio.Unlock()
again:
	if wc.frameReader == nil {
		frame, err := wc.frameReaderFactory.NewFrameReader()
		if err != nil {
			return 0, err
		}
		wc.frameReader, err = wc.frameHandler.HandleFrame(frame)
		if err != nil {
			return 0, err
		}
		if wc.frameReader == nil {
			goto again
		}
	}
	n, err = wc.frameReader.Read(b)
	if err == io.EOF {
		if trailer := wc.frameReader.TrailerReader(); trailer != nil {
			io.Copy(ioutil.Discard, trailer)
		}
		wc.frameReader = nil
		goto again
	}
	return n, err
}

func (wc *WebsocketConn) Write(b []byte) (n int, err error) {
	err = wc.HandShakeHandle()
	if err != nil {
		return 0, err
	}
	wc.wio.Lock()
	defer wc.wio.Unlock()
	w, err := wc.frameWriterFactory.NewFrameWriter(wc.payloadType)
	if err != nil {
		return 0, err
	}
	n, err = w.Write(b)
	w.Close()
	return n, err
}

func (wc *WebsocketConn) Close() (err error) {
	wc.mux.Lock()
	defer wc.mux.Unlock()
	if wc.frameHandler != nil {
		err = wc.frameHandler.WriteClose(wc.defaultCloseStatus)
	}
	err1 := wc.conn.Close()
	if err != nil {
		return err
	}
	return err1
}

func (wc *WebsocketConn) LocalAddr() net.Addr {
	return wc.laddr
}

func (wc *WebsocketConn) RemoteAddr() net.Addr {
	return wc.raddr
}

func (wc *WebsocketConn) SetDeadline(t time.Time) error {
	return wc.conn.SetDeadline(t)
}

func (wc *WebsocketConn) SetReadDeadline(t time.Time) error {
	return wc.conn.SetReadDeadline(t)
}

func (wc *WebsocketConn) SetWriteDeadline(t time.Time) error {
	return wc.conn.SetWriteDeadline(t)
}

func (wc *WebsocketConn) HandShakeHandle() error {
	var err error
	wc.once.Do(func() {
		wc.mux.Lock()
		defer wc.mux.Unlock()
		if wc.tlsCfg != nil {
			wc.conn, err = xtls.TLSUpgrader(wc.tlsCfg, wc.isClient).Upgrade(wc.conn)
			if err != nil {
				return
			}
		}
		err = wc.handShakeHandle()
	})
	return err
}

func newConn(isClient bool, conn net.Conn, cfg *tls.Config) net.Conn {
	wc := &WebsocketConn{
		isClient: isClient,
		conn:     conn,
		tlsCfg:   cfg,
		reader:   nil,
		writer:   nil,
		laddr:    conn.LocalAddr(),
		raddr:    conn.RemoteAddr(),
	}
	if isClient {
		wc.defaultClientCfg()
	}
	return wc
}

func Server(conn net.Conn, cfg *tls.Config) net.Conn {
	return newConn(false, conn, cfg)
}

func Client(conn net.Conn, cfg *tls.Config) net.Conn {
	return newConn(true, conn, cfg)
}
