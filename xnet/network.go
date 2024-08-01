package xnet

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/xnet/xfakehttp"
	"github.com/peakedshout/go-pandorasbox/xnet/xneterr"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"github.com/peakedshout/go-pandorasbox/xnet/xquic"
	"github.com/peakedshout/go-pandorasbox/xnet/xtcp"
	"github.com/peakedshout/go-pandorasbox/xnet/xtls"
	"github.com/peakedshout/go-pandorasbox/xnet/xudp"
	"github.com/peakedshout/go-pandorasbox/xnet/xwebsocket"
	"net"
)

const (
	NetworkTcp   = "tcp"
	NetworkUdp   = "udp"
	NetworkQuic  = "quic"
	NetworkTLS   = "tls"
	NetworkWs    = "ws"
	NetworkWss   = "wss"
	NetworkHttp  = "http"
	NetworkHttps = "https"
)

type MCallbackFunc func(index int, network string) (cfg *tls.Config, isClient bool, err error)

func MakeNetworkUpgrader(callback MCallbackFunc, networkList ...string) (xnetutil.Upgrader, error) {
	upgraderList := make([]xnetutil.Upgrader, 0, len(networkList))
	for i, network := range networkList {
		cfg, isClient, err := callback(i, network)
		if err != nil {
			return nil, err
		}
		upgrader, err := makeNetworkUpgraderOnce(network, cfg, isClient)
		if err != nil {
			return nil, err
		}
		upgraderList = append(upgraderList, upgrader)
	}
	return xnetutil.NewWarpUpgrader(upgraderList...), nil
}

func makeNetworkUpgraderOnce(network string, cfg *tls.Config, isClient bool) (xnetutil.Upgrader, error) {
	switch network {
	case NetworkTcp:
		return xtcp.XUpgrader(cfg, isClient), nil
	case NetworkUdp:
		return xudp.XUpgrader(cfg, isClient), nil
	case NetworkQuic:
		return xquic.XUpgrader(cfg, isClient), nil
	case NetworkTLS:
		if cfg == nil {
			return nil, xneterr.ErrNilTlsConfig.Errorf(network)
		}
		return xtls.XUpgrader(cfg, isClient), nil
	case NetworkWs:
		return xwebsocket.XUpgrader(cfg, isClient), nil
	case NetworkWss:
		if cfg == nil {
			return nil, xneterr.ErrNilTlsConfig.Errorf(network)
		}
		return xwebsocket.XUpgrader(cfg, isClient), nil
	case NetworkHttp:
		return xfakehttp.XUpgrader(cfg, isClient), nil
	case NetworkHttps:
		if cfg == nil {
			return nil, xneterr.ErrNilTlsConfig.Errorf(network)
		}
		return xfakehttp.XUpgrader(cfg, isClient), nil
	default:
		return nil, xneterr.ErrNetworkIsInvalid.Errorf(network)
	}
}

const (
	BaseNetworkTcp  = NetworkTcp
	BaseNetworkUdp  = NetworkUdp
	BaseNetworkQuic = NetworkQuic
)

func IsStdBaseNetwork(network string) bool {
	switch network {
	case BaseNetworkTcp, BaseNetworkUdp:
		return true
	default:
		return false
	}
}

func IsStreamBaseNetwork(network string) bool {
	switch network {
	case BaseNetworkTcp, BaseNetworkQuic:
		return true
	default:
		return false
	}
}

func MakeBaseStreamListener(network string, addr string) (net.Listener, error) {
	switch network {
	case BaseNetworkTcp:
		listen, err := net.Listen(network, addr)
		if err != nil {
			return nil, err
		}
		return listen, nil
	case BaseNetworkQuic:
		listen, err := xquic.Listen(network, addr)
		if err != nil {
			return nil, err
		}
		return listen, nil
	default:
		return nil, xneterr.ErrNetworkIsInvalid.Errorf(fmt.Sprintf("%s is not stream base network", network))
	}
}

func GetStdBaseNetwork(network string) string {
	switch network {
	case BaseNetworkTcp, "tcp4", "tcp6":
		return BaseNetworkTcp
	case NetworkUdp, NetworkQuic, "udp4", "udp6":
		return BaseNetworkUdp
	default:
		return ""
	}
}

func GetStdBaseNetworkToStream(network string) string {
	switch network {
	case BaseNetworkTcp, "tcp4", "tcp6":
		return BaseNetworkTcp
	case NetworkUdp, NetworkQuic, "udp4", "udp6":
		return NetworkQuic
	default:
		return ""
	}
}

func GetBaseStreamDialer(network string) (xnetutil.Dialer, error) {
	switch network {
	case BaseNetworkTcp:
		dialer := &net.Dialer{}
		return dialer, nil
	case BaseNetworkQuic:
		return xquic.NewDialer(nil, nil), nil
	default:
		return nil, xneterr.ErrNetworkIsInvalid.Errorf(fmt.Sprintf("%s is not stream base network", network))
	}
}

func GetBaseStreamListener(network string, addr string) (net.Listener, error) {
	switch network {
	case BaseNetworkTcp:
		lc := net.ListenConfig{}
		return lc.Listen(context.Background(), network, addr)
	case BaseNetworkQuic:
		return xquic.Listen(network, addr)
	default:
		return nil, xneterr.ErrNetworkIsInvalid.Errorf(fmt.Sprintf("%s is not stream base network", network))
	}
}

func GetStreamListener(network string, addr string) (net.Listener, error) {
	network = GetStdBaseNetwork(network)
	switch network {
	case BaseNetworkTcp:
		lc := net.ListenConfig{}
		return lc.Listen(context.Background(), network, addr)
	case BaseNetworkUdp:
		return xquic.Listen(network, addr)
	default:
		return nil, xneterr.ErrNetworkIsInvalid.Errorf(fmt.Sprintf("%s is not stream base network", network))
	}
}

func GetBaseStreamListenerConfig(network string) (xnetutil.ListenerConfig, error) {
	switch network {
	case BaseNetworkTcp:
		return xnetutil.NewDefaultListenerConfig(nil), nil
	case BaseNetworkQuic:
		return xquic.NewQuicListenConfig(nil, nil), nil
	default:
		return nil, xneterr.ErrNetworkIsInvalid.Errorf(fmt.Sprintf("%s is not stream base network", network))
	}
}
