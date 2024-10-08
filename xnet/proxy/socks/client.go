package socks

import (
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
)

func SOCKS4CONNECT(network string, address string, userid S4UserId, forward xnetutil.Dialer) (xnetutil.Dialer, error) {
	return newSocks4Config(network, address, socks4CDCONNECT, userid, forward, nil)
}

func SOCKS4BIND(network string, address string, userid S4UserId, forward xnetutil.Dialer, bindCb BINDAddrCb) (xnetutil.Dialer, error) {
	return newSocks4Config(network, address, socks4CDBIND, userid, forward, bindCb)
}

func SOCKS5CONNECT(network string, address string, auth *S5Auth, forward xnetutil.Dialer) (xnetutil.Dialer, error) {
	return newSocks5Config(network, address, socks5CMDCONNECT, auth, forward, nil, nil, nil)
}
func SOCKS5BIND(network string, address string, auth *S5Auth, forward xnetutil.Dialer, bindCb BINDAddrCb) (xnetutil.Dialer, error) {
	return newSocks5Config(network, address, socks5CMDBIND, auth, forward, nil, bindCb, nil)
}

func SOCKS5UDPASSOCIATE(network string, address string, auth *S5Auth, forward xnetutil.Dialer, uforward xnetutil.PacketListenerConfig, udpCb UDPDataHandler) (xnetutil.PacketListenerConfig, error) {
	return newSocks5Config(network, address, socks5CMDUDPASSOCIATE, auth, forward, uforward, nil, udpCb)
}

func SOCKS5CONNECTP(network string, address string, auth *S5AuthPassword, forward xnetutil.Dialer) (xnetutil.Dialer, error) {
	a := &S5Auth{
		Socks5AuthNOAUTH:   DefaultAuthConnCb,
		Socks5AuthPASSWORD: auth,
	}
	return newSocks5Config(network, address, socks5CMDCONNECT, a, forward, nil, nil, nil)
}
func SOCKS5BINDP(network string, address string, auth *S5AuthPassword, forward xnetutil.Dialer, bindCb BINDAddrCb) (xnetutil.Dialer, error) {
	a := &S5Auth{
		Socks5AuthNOAUTH:   DefaultAuthConnCb,
		Socks5AuthPASSWORD: auth,
	}
	return newSocks5Config(network, address, socks5CMDBIND, a, forward, nil, bindCb, nil)
}

func SOCKS5UDPASSOCIATEP(network string, address string, auth *S5AuthPassword, forward xnetutil.Dialer, uforward xnetutil.PacketListenerConfig, udpCb UDPDataHandler) (xnetutil.PacketListenerConfig, error) {
	a := &S5Auth{
		Socks5AuthNOAUTH:   DefaultAuthConnCb,
		Socks5AuthPASSWORD: auth,
	}
	return newSocks5Config(network, address, socks5CMDUDPASSOCIATE, a, forward, uforward, nil, udpCb)
}
