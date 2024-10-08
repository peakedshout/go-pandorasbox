package xmulti

import (
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/tool/merror"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"github.com/peakedshout/go-pandorasbox/xnet/xquic"
	"net"
	"sync"
)

var DefaultMultiAddrDialer = NewDefaultMultiAddrDialer()

func NewDefaultMultiAddrDialer() *MultiAddrDialer {
	return NewMultiAddrDialer(MultiAddrDialTypeGo, DefaultStreamDialerMap)
}

var DefaultStreamDialerMap = map[string]xnetutil.Dialer{
	"tcp":  &net.Dialer{},
	"udp":  xquic.NewDialer(nil, nil),
	"quic": xquic.NewDialer(nil, nil),
}

type MultiAddrDialType uint8

const (
	MultiAddrDialTypeTurn = MultiAddrDialType(iota)
	MultiAddrDialTypeGo
)

type MultiAddrDialer struct {
	xdm   map[string]xnetutil.Dialer
	sType MultiAddrDialType
}

func NewMultiAddrDialer(sType MultiAddrDialType, drMap map[string]xnetutil.Dialer) *MultiAddrDialer {
	return &MultiAddrDialer{
		xdm:   drMap,
		sType: sType,
	}
}

func (md *MultiAddrDialer) Dial(network string, address string) (conn net.Conn, err error) {
	return md.DialContext(context.Background(), network, address)
}

func (md *MultiAddrDialer) DialContext(ctx context.Context, network string, address string) (conn net.Conn, err error) {
	return md.MultiDialContext(ctx, xnetutil.NewNetAddr(network, address))
}

func (md *MultiAddrDialer) MultiDial(addr ...net.Addr) (conn net.Conn, err error) {
	return md.MultiDialContext(context.Background(), addr...)
}

func (md *MultiAddrDialer) MultiDialContext(ctx context.Context, addr ...net.Addr) (conn net.Conn, err error) {
	if len(addr) == 0 {
		return nil, errors.New("nil dial address")
	} else if len(addr) == 1 {
		dr, err := md.getDialer(addr[0].Network())
		if err != nil {
			return nil, err
		}
		conn, err = dr.DialContext(ctx, addr[0].Network(), addr[0].String())
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	switch md.sType {
	case MultiAddrDialTypeTurn:
		multiErr := merror.NewMultiErr("multiAddrDialer err")
		for _, s := range addr {
			dr, err := md.getDialer(s.Network())
			if err != nil {
				if ctxtool.Disable(ctx) {
					return nil, err
				}
				multiErr.AddErr(s.String(), err)
				continue
			}
			conn, err = dr.DialContext(ctx, s.Network(), s.String())
			if err != nil {
				if ctxtool.Disable(ctx) {
					return nil, err
				}
				multiErr.AddErr(s.String(), err)
				continue
			}
			return conn, nil
		}
		return nil, multiErr
	case MultiAddrDialTypeGo:
		multiErr := merror.NewMultiErr("multiAddrDialer err")
		nctx, cancelFunc := context.WithCancel(ctx)
		defer cancelFunc()
		var mux sync.Mutex
		var wg sync.WaitGroup
		for _, s := range addr {
			one := s
			dr, err := md.getDialer(one.Network())
			if err != nil {
				multiErr.AddErr(one.String(), err)
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				gconn, gerr := dr.DialContext(nctx, one.Network(), one.String())
				if gerr != nil {
					if ctxtool.Disable(nctx) {
						return
					}
					multiErr.AddErr(one.String(), gerr)
				}
				mux.Lock()
				defer mux.Unlock()
				if conn == nil {
					cancelFunc()
					conn = gconn
				} else {
					gconn.Close()
				}
			}()
		}
		wg.Wait()
		if conn != nil {
			return conn, nil
		}
		return nil, multiErr
	default:
		return nil, errors.New("MultiAddrDialType invalid")
	}
}

func (md *MultiAddrDialer) getDialer(network string) (xnetutil.Dialer, error) {
	dr := md.xdm[network]
	if dr == nil {
		return nil, fmt.Errorf("nil %s network dialer", network)
	}
	return dr, nil
}
