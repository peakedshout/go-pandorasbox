package xmulti

import (
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/merror"
	"github.com/peakedshout/go-pandorasbox/xnet/xneterr"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"github.com/peakedshout/go-pandorasbox/xnet/xquic"
	"net"
	"sync"
)

type multiListener struct {
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan net.Conn
	addr   net.Addr
	once   sync.Once
	xl     []net.Listener
	bind   bool
}

func NewMultiListenerFromAddr(ctx context.Context, bind bool, addr ...net.Addr) (net.Listener, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	merr := merror.NewMultiErr("multiListenerFromAddr err")
	lns := make([]net.Listener, 0, len(addr))
	for _, n := range addr {
		switch n.Network() {
		case "tcp", "tcp4", "tcp6":
			listener, err := xnetutil.NewDefaultListenerConfig(nil).ListenContext(ctx, n.Network(), n.String())
			if err != nil {
				merr.AddErr(fmt.Sprintf("%s_%s", n.Network(), n.String()), err)
			} else {
				lns = append(lns, listener)
			}
		case "udp", "udp4", "udp6", "quic":
			listener, err := xquic.ListenContext(ctx, n.Network(), n.String())
			if err != nil {
				merr.AddErr(fmt.Sprintf("%s_%s", n.Network(), n.String()), err)
			} else {
				lns = append(lns, listener)
			}
		default:
			merr.AddErr(fmt.Sprintf("%s_%s", n.Network(), n.String()), xneterr.ErrNetworkIsInvalid.Errorf(fmt.Sprintf("%s is not stream base network", n.Network())))
		}
	}
	if !merr.Nil() {
		for _, ln := range lns {
			ln.Close()
		}
		return nil, merr
	}
	return NewMultiListener(ctx, bind, lns...)
}

func NewMultiListener(ctx context.Context, bind bool, lns ...net.Listener) (net.Listener, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ml := &multiListener{
		ch:   make(chan net.Conn),
		addr: nil,
		once: sync.Once{},
		xl:   lns,
		bind: bind,
	}
	ml.ctx, ml.cancel = context.WithCancel(ctx)
	switch len(ml.xl) {
	case 0:
		return nil, errors.New("nil listener unit")
	case 1:
		ml.addr = ml.xl[0].Addr()
	default:
		al := make(xnetutil.MultiAddr, 0, len(ml.xl))
		for _, listener := range ml.xl {
			al = append(al, listener.Addr())
		}
		ml.addr = al
	}
	return ml, nil
}

func (ml *multiListener) Accept() (net.Conn, error) {
	if len(ml.xl) <= 1 {
		return ml.xl[0].Accept()
	} else {
		ml.once.Do(func() {
			go ml.run(ml.bind, ml.xl...)
		})
		select {
		case conn := <-ml.ch:
			return conn, nil
		case <-ml.ctx.Done():
			return nil, ml.ctx.Err()
		}
	}
}

func (ml *multiListener) Close() error {
	if len(ml.xl) <= 1 {
		return ml.xl[0].Close()
	}
	ml.cancel()
	ml.multiClose()
	err := ml.ctx.Err()
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (ml *multiListener) Addr() net.Addr {
	return ml.addr
}

func (ml *multiListener) run(bind bool, lns ...net.Listener) {
	defer ml.cancel()
	var once sync.Once
	closeFn := func() {
		once.Do(ml.multiClose)
	}
	defer closeFn()
	var wg sync.WaitGroup
	wg.Add(len(lns))
	for _, ln := range lns {
		lnr := ln
		go func() {
			defer func() {
				wg.Done()
				if bind {
					closeFn()
				}
			}()
			for {
				conn, err := lnr.Accept()
				if err != nil {
					return
				}
				select {
				case ml.ch <- conn:
				case <-ml.ctx.Done():
					return
				}
			}
		}()
	}
	wg.Wait()
}

func (ml *multiListener) multiClose() {
	for _, ln := range ml.xl {
		ln.Close()
	}
}
