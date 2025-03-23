package xcmd

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/xnet"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
	"io"
	"net"
)

type Auth struct {
	UserName string `json:"userName" yaml:"userName"`
	Password string `json:"password" yaml:"password"`
}

type TlsType string

const (
	TlsTypeDefault  TlsType = "default"
	TlsTypeEnable   TlsType = "enable"
	TlsTypeInsecure TlsType = "insecure"
	TlsTypeDisable  TlsType = "disable"
)

type Config struct {
	Type       xhttp.Type `json:"type" yaml:"type"`
	Auth       *Auth      `json:"auth" yaml:"auth"`
	Prefix     string     `json:"prefix" yaml:"prefix"`
	TlsType    TlsType    `json:"tlsType" yaml:"tlsType"`
	Cert, Key  []byte     `json:"-" yaml:"-"`
	CertFile   string     `json:"certFile" yaml:"certFile"`
	KeyFile    string     `json:"keyFile" yaml:"keyFile"`
	Network    string     `json:"network" yaml:"network"`
	Address    string     `json:"address" yaml:"address"`
	OnlyClient bool       `json:"onlyClient" yaml:"onlyClient"`

	Ln net.Listener    `json:"-" yaml:"-"`
	Dr xnetutil.Dialer `json:"-" yaml:"-"`
}

type XCmd struct {
	ctx  context.Context
	cl   context.CancelFunc
	s    *xhttp.Server
	c    *xhttp.Client
	n, a string
	ln   net.Listener
}

func NewXCmd(ctx context.Context, cfg *Config) (*XCmd, error) {
	var tcfg *tls.Config
	if cfg.TlsType != TlsTypeDisable {
		if cfg.Cert != nil {
			tc, err := pcrypto.MakeTlsConfig(cfg.Cert, cfg.Key)
			if err != nil {
				return nil, err
			}
			tcfg = tc
		} else if cfg.CertFile != "" {
			tc, err := pcrypto.MakeTlsConfigFromFile(cfg.CertFile, cfg.KeyFile)
			if err != nil {
				return nil, err
			}
			tcfg = tc
		}
	}
	switch cfg.TlsType {
	case TlsTypeEnable, TlsTypeInsecure:
		if !cfg.OnlyClient && tcfg == nil {
			return nil, errors.New("nil tls config")
		}
		if tcfg == nil {
			tcfg = new(tls.Config)
		}
		if cfg.TlsType == TlsTypeInsecure {
			tcfg.InsecureSkipVerify = true
		}
	case TlsTypeDisable:
		tcfg = nil
	default:

	}

	xctx, xcl := context.WithCancel(ctx)
	c := &xhttp.Config{
		Ctx:    xctx,
		Type:   0,
		TlsCfg: tcfg,
		Auth:   nil,
		Prefix: cfg.Prefix,
	}
	var dr xnetutil.Dialer
	if cfg.Dr != nil {
		dr = cfg.Dr
	} else {
		dialer, err := xnet.GetBaseStreamDialer(xnet.GetStdBaseNetworkToStream(cfg.Network))
		if err != nil {
			xcl()
			return nil, err
		}
		dr = dialer
	}
	cc := &xhttp.ClientConfig{
		Dr:     dr,
		TlsCfg: tcfg,
		Prefix: cfg.Prefix,
		Host:   cfg.Address,
	}
	if cfg.Auth != nil {
		c.Auth = func(u, p string) bool {
			if u != cfg.Auth.UserName || p != cfg.Auth.Password {
				return false
			}
			return true
		}
		cc.Auth = &struct {
			AuthUserName string
			AuthPassword string
		}{AuthUserName: cfg.Auth.UserName, AuthPassword: cfg.Auth.Password}
	}
	x := &XCmd{
		ctx: xctx,
		cl:  xcl,
		c:   xhttp.NewClient(cc),
		n:   cfg.Network,
		a:   cfg.Address,
		ln:  cfg.Ln,
	}
	if !cfg.OnlyClient {
		x.s = xhttp.NewServer(c)
	}
	return x, nil
}

func (x *XCmd) Call(ctx context.Context, name string, data any) (io.ReadCloser, error) {
	return x.c.Call(ctx, name, data)
}

func (x *XCmd) CallBytes(ctx context.Context, name string, data any) ([]byte, error) {
	return x.c.CallBytes(ctx, name, data)
}

func (x *XCmd) CallAny(ctx context.Context, name string, data any, out any) error {
	return x.c.CallAny(ctx, name, data, out)
}

func (x *XCmd) Client() *xhttp.Client {
	return x.c
}

func (x *XCmd) Serve() (err error) {
	if x.s == nil {
		return errors.New("only client mode")
	}
	ln := x.ln
	if ln == nil {
		ln, err = xnet.GetStreamListener(x.n, x.a)
		if err != nil {
			return err
		}
	}
	return x.s.AutoServe(ln)
}

func (x *XCmd) AsyncServe() (err error) {
	if x.s == nil {
		return errors.New("only client mode")
	}
	ln := x.ln
	if ln == nil {
		ln, err = xnet.GetStreamListener(x.n, x.a)
		if err != nil {
			return err
		}
	}
	go x.s.AutoServe(ln)
	return nil
}

func (x *XCmd) AsyncServeCallBack(fn func(err error)) (err error) {
	if x.s == nil {
		return errors.New("only client mode")
	}
	ln := x.ln
	if ln == nil {
		ln, err = xnet.GetStreamListener(x.n, x.a)
		if err != nil {
			return err
		}
	}
	go func() {
		fn(x.s.AutoServe(ln))
	}()
	return nil
}

func (x *XCmd) Close() error {
	x.cl()
	x.c.Close()
	if x.s != nil {
		return x.s.Close()
	}
	return nil
}

func (x *XCmd) Set(name string, handlers ...xhttp.Handler) {
	if x.s != nil {
		x.s.Set(name, handlers...)
	}
}

func (x *XCmd) Context() context.Context {
	return x.ctx
}
