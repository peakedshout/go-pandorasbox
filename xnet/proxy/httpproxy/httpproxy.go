package httpproxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func UserInfoAuth(username string, password string) ReqAuthCb {
	return func(req *http.Request) bool {
		u, p, ok := proxyBasicAuth(req)
		if ok && u == username && p == password {
			return true
		} else {
			return false
		}
	}
}

func Serve(ln net.Listener, cfg *ServerConfig) error {
	server, err := NewServer(cfg)
	if err != nil {
		return err
	}
	defer server.Close()
	return server.Serve(ln)
}

func ListenAndServe(network string, addr string, cfg *ServerConfig) error {
	server, err := NewServer(cfg)
	if err != nil {
		return err
	}
	defer server.Close()
	return server.ListenAndServe(network, addr)
}

type ReqAuthCb func(req *http.Request) bool

type ServerConfig struct {
	ReqAuthCb   ReqAuthCb
	Forward     xnetutil.Dialer
	DialTimeout time.Duration
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	return NewServerContext(context.Background(), cfg)
}

func NewServerContext(ctx context.Context, cfg *ServerConfig) (*Server, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg == nil {
		return nil, errors.New("nil server config")
	}
	s := &Server{
		cfg: cfg,
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	return s, nil
}

type Server struct {
	cfg *ServerConfig

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Server) Serve(ln net.Listener) error {
	return s.listen(ln)
}

func (s *Server) ListenAndServe(network string, addr string) error {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	return s.Serve(ln)
}

func (s *Server) Close() error {
	s.cancel()
	return s.ctx.Err()
}

func (s *Server) listen(ln net.Listener) error {
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return
	}

	if s.cfg.ReqAuthCb != nil && !s.cfg.ReqAuthCb(req) {
		resp := http.Response{
			Status:     "401 Unauthorized",
			StatusCode: http.StatusUnauthorized,
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
		}
		_ = resp.Write(conn)
		return
	}

	otherMethod := false
	bs := new(bytes.Buffer)
	if req.Method != http.MethodConnect {
		err = req.Write(bs)
		if err != nil {
			return
		}
		otherMethod = true
	}

	host := req.Host
	_, _, err = net.SplitHostPort(host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			host += ":80"
		} else {
			return
		}
	}
	ctx := s.ctx
	if s.cfg.DialTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.cfg.DialTimeout)
		defer cancel()
	}
	var rconn net.Conn
	if s.cfg.Forward != nil {
		rconn, err = s.cfg.Forward.DialContext(s.ctx, "tcp", host)
		if err != nil {
			return
		}
	} else {
		dr := &net.Dialer{}
		rconn, err = dr.DialContext(s.ctx, "tcp", host)
		if err != nil {
			return
		}
	}
	defer rconn.Close()

	// http need last req
	if otherMethod {
		_, err = rconn.Write(bs.Bytes())
		if err != nil {
			return
		}
	} else {
		// https will
		resp := http.Response{
			Status:     "200 Connection Established",
			StatusCode: 200,
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
		}
		err = resp.Write(conn)
		if err != nil {
			return
		}
	}

	go io.Copy(conn, rconn)
	io.Copy(rconn, conn)
}

func proxyBasicAuth(req *http.Request) (username, password string, ok bool) {
	username, password, ok = ProxyBasicAuth(req)
	if ok {
		req.Header.Del("Proxy-Authorization")
	}
	return username, password, ok
}

func ProxyBasicAuth(req *http.Request) (username, password string, ok bool) {
	auth := req.Header.Get("Proxy-Authorization")
	if auth == "" {
		return "", "", false
	}
	return parseBasicAuth(auth)
}

func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(auth) < len(prefix) || !equalFold(auth[:len(prefix)], prefix) {
		return "", "", false
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}

func equalFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if lower(s[i]) != lower(t[i]) {
			return false
		}
	}
	return true
}

// lower returns the ASCII lowercase version of b.
func lower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
