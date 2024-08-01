package fakehttpconn

import (
	"bufio"
	"crypto/tls"
	"errors"
	"github.com/peakedshout/go-pandorasbox/ccw/closer"
	"github.com/peakedshout/go-pandorasbox/tool/bio"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type FakeHttpConn struct {
	isClient bool
	conn     net.Conn
	cfg      *tls.Config
	reader   *bufio.Reader
	readBody io.ReadCloser
	readLock sync.Mutex
	readable bool
	laddr    net.Addr
	raddr    net.Addr
	closer   closer.Closer

	clientTemplate http.Request
	serverTemplate http.Response
}

func Server(conn net.Conn, cfg *tls.Config) net.Conn {
	return newFakeHttpConn(false, conn, cfg)
}

func Client(conn net.Conn, cfg *tls.Config) net.Conn {
	return newFakeHttpConn(true, conn, cfg)
}

func newFakeHttpConn(isClient bool, conn net.Conn, cfg *tls.Config) net.Conn {
	fhc := &FakeHttpConn{
		isClient: isClient,
		conn:     conn,
		cfg:      cfg,
		closer:   closer.NewCloser(),
	}
	if fhc.cfg != nil {
		if fhc.isClient {
			fhc.conn = tls.Client(fhc.conn, fhc.cfg)
		} else {
			fhc.conn = tls.Server(fhc.conn, fhc.cfg)
		}
	}
	fhc.laddr = fhc.conn.LocalAddr()
	fhc.raddr = fhc.conn.RemoteAddr()
	fhc.reader = bufio.NewReader(fhc.conn)
	fhc.makeTemplate()
	fhc.closer.AddCloseFn(func() {
		fhc.conn.Close()
		if fhc.readBody != nil {
			fhc.readBody.Close()
		}
	})
	return fhc
}

func (fhc *FakeHttpConn) Read(b []byte) (n int, err error) {
	return fhc.read(b)
}

func (fhc *FakeHttpConn) Write(b []byte) (n int, err error) {
	return fhc.write(b)
}

func (fhc *FakeHttpConn) Close() error {
	return fhc.closer.Close()
}

func (fhc *FakeHttpConn) LocalAddr() net.Addr  { return fhc.laddr }
func (fhc *FakeHttpConn) RemoteAddr() net.Addr { return fhc.raddr }

func (fhc *FakeHttpConn) SetDeadline(t time.Time) error {
	return fhc.conn.SetDeadline(t)
}
func (fhc *FakeHttpConn) SetReadDeadline(t time.Time) error {
	return fhc.conn.SetReadDeadline(t)
}
func (fhc *FakeHttpConn) SetWriteDeadline(t time.Time) error {
	return fhc.conn.SetWriteDeadline(t)
}

func (fhc *FakeHttpConn) read(b []byte) (n int, err error) {
	fhc.readLock.Lock()
	defer fhc.readLock.Unlock()
	for {
		if fhc.readable && fhc.readBody != nil {
			n, err = fhc.readBody.Read(b)
			if err != nil {
				if errors.Is(err, io.EOF) {
					fhc.readBody.Close()
					fhc.readable = false
					if n == 0 {
						continue
					} else {
						err = nil
					}
				}
			}
			return n, err
		} else {
			if fhc.isClient {
				err = fhc.readClient()
			} else {
				err = fhc.readServer()
			}
			if err != nil {
				return 0, err
			}
			fhc.readable = true
		}
	}
}

func (fhc *FakeHttpConn) readClient() error {
	for {
		resp, err := http.ReadResponse(fhc.reader, nil)
		if err != nil {
			return err
		}
		if resp.Body == nil {
			continue
		}
		fhc.readBody = resp.Body
		return nil
	}
}

func (fhc *FakeHttpConn) readServer() error {
	for {
		req, err := http.ReadRequest(fhc.reader)
		if err != nil {
			return err
		}
		if req.Body == nil {
			continue
		}
		fhc.readBody = req.Body
		return nil
	}
}

func (fhc *FakeHttpConn) write(b []byte) (n int, err error) {
	if fhc.isClient {
		return fhc.writeClient(b)
	} else {
		return fhc.writeServer(b)
	}
}

func (fhc *FakeHttpConn) writeClient(b []byte) (n int, err error) {
	req := fhc.clientTemplate
	l := len(b)
	req.Body = bio.NewNoCloseBody(b)
	req.ContentLength = int64(l)
	err = req.Write(fhc.conn)
	if err != nil {
		return 0, err
	}
	return l, nil
}

func (fhc *FakeHttpConn) writeServer(b []byte) (n int, err error) {
	resp := fhc.serverTemplate
	l := len(b)
	resp.Body = bio.NewNoCloseBody(b)
	resp.ContentLength = int64(l)
	err = resp.Write(fhc.conn)
	if err != nil {
		return 0, err
	}
	return l, nil
}

func (fhc *FakeHttpConn) makeTemplate() {
	if fhc.isClient {
		u, _ := url.Parse(fhc.raddr.Network() + "://" + fhc.raddr.String())
		fhc.clientTemplate = http.Request{
			Method:        http.MethodGet,
			URL:           u,
			ContentLength: 0,
			Host:          fhc.raddr.String(),
		}
	} else {
		fhc.serverTemplate = http.Response{
			StatusCode:    http.StatusOK,
			ProtoMajor:    1,
			ProtoMinor:    1,
			ContentLength: 0,
			Request:       &http.Request{Method: http.MethodGet},
		}
	}
}
