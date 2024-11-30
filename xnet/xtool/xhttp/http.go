package xhttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"sync/atomic"
)

type Type uint8

const (
	TypeNone Type = iota
	TypeMuxReject
	TypeMuxWait
)

const (
	ArgsQuery  = "xhttp-query"
	ArgsHeader = "xhttp-header"
)

type Config struct {
	Ctx    context.Context
	Type   Type
	TlsCfg *tls.Config
	Auth   func(u, p string) bool
	Prefix string
}

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	hs     *http.Server
	hm     *http.ServeMux
	prefix string

	auth func(u, p string) bool

	// serve
	sType   Type
	waitCh  chan struct{}
	calling *atomic.Bool
}

func NewServer(cfg *Config) *Server {
	if cfg.Ctx == nil {
		cfg.Ctx = context.Background()
	}
	pctx, pcancel := signal.NotifyContext(cfg.Ctx, os.Interrupt, os.Kill)
	s := &Server{
		ctx:    pctx,
		cancel: pcancel,
		hs: &http.Server{BaseContext: func(listener net.Listener) context.Context {
			return pctx
		}, TLSConfig: cfg.TlsCfg, ErrorLog: log.New(io.Discard, "", log.LstdFlags)},
		hm:      nil,
		prefix:  cfg.Prefix,
		auth:    cfg.Auth,
		sType:   cfg.Type,
		waitCh:  make(chan struct{}, 1),
		calling: &atomic.Bool{},
	}
	s.hm = http.NewServeMux()
	s.hs.Handler = s.hm
	ctxtool.GWaitFunc(pctx, func() {
		_ = s.Close()
	})
	return s
}

// AutoServe if had tls config will be used
func (s *Server) AutoServe(ln net.Listener) error {
	if s.hs.TLSConfig != nil {
		ln = tls.NewListener(ln, s.hs.TLSConfig)
	}
	return s.hs.Serve(ln)
}

func (s *Server) Serve(ln net.Listener) error {
	return s.hs.Serve(ln)
}

func (s *Server) ServeTLS(l net.Listener, certFile string, keyFile string) error {
	return s.hs.ServeTLS(l, certFile, keyFile)
}

func (s *Server) Close() error {
	s.cancel()
	return s.hs.Close()
}

func (s *Server) Set(name string, handlers ...Handler) {
	s.hm.HandleFunc(path.Join("/", s.prefix, name), func(writer http.ResponseWriter, request *http.Request) {
		if !s.authCb(writer, request) {
			return
		}
		fn, ok := s.muxCb(writer, request)
		if !ok {
			return
		}
		defer fn()
		ctx := &Context{
			Context: request.Context(),
			url:     request.URL,
			h:       request.Header,
			r:       request.Body,
			w:       writer,
		}
		for _, handler := range handlers {
			err := handler(ctx)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(err.Error()))
				return
			}
		}
	})
}

func (s *Server) Context() context.Context {
	return s.ctx
}

func (s *Server) authCb(w http.ResponseWriter, r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if s.auth != nil {
		if !ok || !s.auth(username, password) {
			_, _ = w.Write([]byte(ErrAuthFailed.Error()))
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
	}
	return true
}

func (s *Server) muxCb(w http.ResponseWriter, r *http.Request) (fn func(), ok bool) {
	defer func() {
		if !ok {
			_, _ = w.Write([]byte(ErrRejectHandle.Error()))
			w.WriteHeader(http.StatusBadRequest)
		}
	}()
	switch s.sType {
	case TypeMuxReject:
		if s.calling.Swap(true) {
			return nil, false
		}
		return func() {
			s.calling.Store(false)
		}, true
	case TypeMuxWait:
		select {
		case s.waitCh <- struct{}{}:
			return func() {
				<-s.waitCh
			}, true
		case <-r.Context().Done():
			return nil, false
		}
	default:
		return func() {}, true
	}
}

var ErrAuthFailed = errors.New("auth failed")
var ErrRejectHandle = errors.New("reject handle")

type Handler func(*Context) error

type Context struct {
	context.Context
	url *url.URL
	h   http.Header
	r   io.ReadCloser
	w   http.ResponseWriter
}

func (c *Context) Query() url.Values {
	return c.url.Query()
}

func (c *Context) RHeader() http.Header {
	return c.h
}

func (c *Context) WHeader() http.Header {
	return c.w.Header()
}

func (c *Context) Bind(a any) error {
	if c.r == nil {
		return errors.New("nil body")
	}
	if v, ok := a.(*io.ReadCloser); ok {
		if v == nil {
			return errors.New("nil ptr")
		}
		*v = c.r
		return nil
	}
	defer c.r.Close()
	return json.NewDecoder(c.r).Decode(a)
}

func (c *Context) WriteFlush(b []byte) (int, error) {
	flusher, ok := c.w.(http.Flusher)
	if !ok {
		return 0, errors.New("not flusher")
	}
	n, err := c.w.Write(b)
	if err != nil {
		return n, err
	}
	flusher.Flush()
	return n, nil
}

func (c *Context) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

func (c *Context) WriteString(s string) (int, error) {
	return c.w.Write([]byte(s))
}

func (c *Context) WriteLn(b []byte) (int, error) {
	return c.w.Write(append(b, '\n'))
}

func (c *Context) WriteStringLn(s string) (int, error) {
	return c.w.Write(append([]byte(s), '\n'))
}

func (c *Context) WriteAny(a any) error {
	c.w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(c.w).Encode(a)
}

type ClientConfig struct {
	Dr     xnetutil.Dialer
	TlsCfg *tls.Config
	Auth   *struct {
		AuthUserName string
		AuthPassword string
	}
	Prefix string
	Host   string
}

type Client struct {
	closed atomic.Bool
	c      *http.Client
	auth   *struct {
		u string
		p string
	}
	prefix string
	host   string
	s      string
}

func NewClient(cfg *ClientConfig) *Client {
	ht := &http.Transport{
		TLSClientConfig: cfg.TlsCfg,
	}
	if cfg.Dr != nil {
		ht.DialContext = cfg.Dr.DialContext
	}
	s := "http"
	if cfg.TlsCfg != nil {
		s = "https"
	}
	c := &Client{
		c: &http.Client{
			Transport: ht,
		},
		prefix: cfg.Prefix,
		host:   cfg.Host,
		s:      s,
	}
	if cfg.Auth != nil {
		c.auth = &struct {
			u string
			p string
		}{u: cfg.Auth.AuthUserName, p: cfg.Auth.AuthPassword}
	}
	if c.host == "" {
		c.host = "127.0.0.1:" + c.s
	}
	return c
}

func (c *Client) Close() {
	c.closed.Store(true)
	c.c.CloseIdleConnections()
}

func (c *Client) Call(ctx context.Context, name string, data any) (io.ReadCloser, error) {
	response, err := c.call(ctx, name, data)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		all, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		msg := response.Status
		if len(all) != 0 {
			msg = string(all)
		}
		return nil, fmt.Errorf("[%d]%s", response.StatusCode, msg)
	}
	return response.Body, nil
}

func (c *Client) CallBytes(ctx context.Context, name string, data any) ([]byte, error) {
	response, err := c.call(ctx, name, data)
	if err != nil {
		return nil, err
	}
	all, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		msg := response.Status
		if len(all) != 0 {
			msg = string(all)
		}
		return nil, fmt.Errorf("[%d]%s", response.StatusCode, msg)
	}
	return all, nil
}

func (c *Client) CallAny(ctx context.Context, name string, data, out any) error {
	bs, err := c.CallBytes(ctx, name, data)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(bs, out)
}

func (c *Client) getUrl(ctx context.Context, name string) string {
	query := getArgsQuery(ctx).Encode()
	if query == "" {
		return c.s + "://" + path.Join(c.host, c.prefix, name)
	}
	return c.s + "://" + path.Join(c.host, c.prefix, name+"?"+query)
}

func (c *Client) call(ctx context.Context, name string, data any) (*http.Response, error) {
	if c.closed.Load() {
		return nil, errors.New("client closed")
	}
	var r io.Reader
	if data != nil {
		if reader, ok := data.(io.Reader); ok {
			r = reader
		} else {
			bs := new(bytes.Buffer)
			err := json.NewEncoder(bs).Encode(data)
			if err != nil {
				return nil, err
			}
			r = bs
		}
	}
	req, err := http.NewRequestWithContext(ctx, "", c.getUrl(ctx, name), r)
	if err != nil {
		return nil, err
	}
	header := getArgsHeader(ctx)
	if header != nil {
		req.Header = header
	}
	if c.auth != nil {
		req.SetBasicAuth(c.auth.u, c.auth.p)
	}
	response, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getArgsQuery(ctx context.Context) url.Values {
	value := ctx.Value(ArgsQuery)
	if value == nil {
		return nil
	}
	uv, ok := value.(url.Values)
	if ok {
		return uv
	}
	return nil
}

func getArgsHeader(ctx context.Context) http.Header {
	value := ctx.Value(ArgsHeader)
	if value == nil {
		return nil
	}
	uv, ok := value.(http.Header)
	if ok {
		return uv
	}
	return nil
}

func SetArgsQuery(ctx context.Context, values url.Values) context.Context {
	return context.WithValue(ctx, ArgsQuery, values)
}

func SetArgsHeader(ctx context.Context, header http.Header) context.Context {
	return context.WithValue(ctx, ArgsHeader, header)
}
