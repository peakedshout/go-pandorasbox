package xnetutil

import (
	"fmt"
	"github.com/peakedshout/go-pandorasbox/xnet/xneterr"
	"net"
	"reflect"
)

func TypeCheck(conn net.Conn, tmp ...net.Conn) (int, error) {
	tf := reflect.TypeOf(conn)
	tn := make([]string, 0, len(tmp))
	for i, n := range tmp {
		ntf := reflect.TypeOf(n)
		if tf.AssignableTo(ntf) {
			return i, nil
		}
		tn = append(tn, ntf.String())
	}
	return -1, xneterr.ErrConnTypeIsInvalid.Errorf(tn)
}

type simpleAddr struct {
	network string
	address string
}

func (s *simpleAddr) Network() string {
	return s.network
}

func (s *simpleAddr) String() string {
	return s.address
}

func NewNetAddr(network, address string) net.Addr {
	return &simpleAddr{
		network: network,
		address: address,
	}
}

type MultiAddr []net.Addr

func (ma MultiAddr) Network() string {
	sl := make([]string, 0, len(ma))
	for _, addr := range ma {
		sl = append(sl, addr.Network())
	}
	return fmt.Sprint(sl)
}

func (ma MultiAddr) String() string {
	sl := make([]string, 0, len(ma))
	for _, addr := range ma {
		sl = append(sl, addr.Network()+"_"+addr.String())
	}
	return fmt.Sprint(sl)
}
