//go:build windows

package control

import (
	"errors"
	"golang.org/x/sys/windows"
	"net/netip"
	"strings"
	"syscall"
)

type Sockaddr interface {
	windows.Sockaddr
	String() string
}

type sockaddr struct {
	windows.Sockaddr
}

func toSockaddr(addr windows.Sockaddr) *sockaddr {
	return &sockaddr{Sockaddr: addr}
}

func (s *sockaddr) String() string {
	return addrToString(s)
}

type FdPtr = windows.Handle

func portReuseControl(network string, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		portReuse(FdPtr(fd))
	})
}

func portReuse(fd FdPtr) error {
	err := windows.SetsockoptInt(fd, windows.SOL_SOCKET, windows.SO_REUSEADDR, 1)
	if err != nil {
		return err
	}
	return nil
}

func rawSockSendTo(fd *FdPtr, b []byte, network, raddr string) (err error) {
	addr, v6, err := parseAddr(parseNetworkToIP(network), raddr)
	if err != nil {
		return err
	}
	if *fd == 0 {
		proto := parseNetworkToProto(network)
		if proto == windows.IPPROTO_IP {
			proto = parseIPToProto(v6)
		}
		sfd, err := windows.Socket(parseIPToDomain(v6), windows.SOCK_RAW, proto)
		if err != nil {
			return err
		}
		*fd = sfd
	}
	return sendTo(*fd, b, addr)
}

func rawSockRecvFrom(fd *FdPtr, b []byte, network string) (n int, raddr Sockaddr, err error) {
	if *fd == 0 {
		v6 := parseNetworkToIP(network)
		proto := parseNetworkToProto(network)
		if proto == windows.IPPROTO_IP {
			proto = parseIPToProto(v6)
		}
		sfd, err := windows.Socket(parseIPToDomain(v6), windows.SOCK_RAW, proto)
		if err != nil {
			return 0, nil, err
		}
		*fd = sfd
	}
	return recvFrom(*fd, b)
}

func newSocket(domain, typ, proto int) (FdPtr, error) {
	sfd, err := windows.Socket(domain, typ, proto)
	if err != nil {
		return 0, err
	}
	return sfd, nil
}

func bind(fd FdPtr, addr Sockaddr) error {
	return windows.Bind(fd, addr)
}

func toClose(fd FdPtr) error {
	if fd == 0 {
		return nil
	}
	return windows.Close(fd)
}

func recvFrom(fd FdPtr, b []byte) (n int, raddr Sockaddr, err error) {
	n, rsd, err := windows.Recvfrom(fd, b, 0)
	if err != nil {
		return 0, nil, err
	}
	return n, toSockaddr(rsd), nil
}

func sendTo(fd FdPtr, b []byte, raddr Sockaddr) error {
	return windows.Sendto(fd, b, 0, raddr)
}

func send(fd FdPtr, b []byte) error {
	return errors.New("not support")
}

func write(fd FdPtr, b []byte) (n int, err error) {
	return windows.Write(fd, b)
}

func read(fd FdPtr, b []byte) (n int, err error) {
	return windows.Read(fd, b)
}

func connect(fd FdPtr, addr Sockaddr) error {
	return windows.Connect(fd, addr)
}

func listen(fd FdPtr, n int) error {
	return windows.Listen(fd, n)
}

func accept(fd FdPtr) (FdPtr, Sockaddr, error) {
	rfd, raddr, err := windows.Accept(fd)
	if err != nil {
		return 0, nil, err
	}
	return rfd, toSockaddr(raddr), nil
}

func parseAddr(v6 bool, addr string) (Sockaddr, bool, error) {
	ap, err := netip.ParseAddrPort(addr)
	if err != nil {
		return nil, false, err
	}
	if ap.Addr().Is4() && !v6 {
		var addr4 [4]byte
		if ap.Addr().IsUnspecified() {
			addr4 = [4]byte{0, 0, 0, 0}
		} else {
			addr4 = ap.Addr().As4()
		}
		a4 := &windows.SockaddrInet4{
			Port: int(ap.Port()),
			Addr: addr4,
		}
		return toSockaddr(a4), false, nil
	} else {
		a6 := &windows.SockaddrInet6{
			Port:   int(ap.Port()),
			ZoneId: 0,
			Addr:   ap.Addr().As16(),
		}
		return toSockaddr(a6), true, nil
	}
}

func parseNetworkToIP(network string) bool {
	if network == "tcp6" || network == "udp6" || strings.HasPrefix(network, "ip6") {
		return true
	} else {
		return false
	}
}

func parseNetworkToType(network string) int {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return windows.SOCK_STREAM
	case "udp", "udp4", "udp6":
		return windows.SOCK_DGRAM
	default:
		return windows.SOCK_RAW
	}
}

func parseIPToDomain(v6 bool) int {
	if v6 {
		return windows.AF_INET6
	} else {
		return windows.AF_INET
	}
}

func parseIPToProto(v6 bool) int {
	if v6 {
		return windows.IPPROTO_IP
	} else {
		return windows.IPPROTO_IPV6
	}
}

func parseNetworkToProto(network string) int {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return windows.IPPROTO_TCP
	case "udp", "udp4", "udp6":
		return windows.IPPROTO_UDP
	default:
		return windows.IPPROTO_IP
	}
}

func addrToString(addr *sockaddr) string {
	if addr == nil || addr.Sockaddr == nil {
		return "nil"
	}
	var ap netip.Addr
	var port int
	switch addr.Sockaddr.(type) {
	case *windows.SockaddrInet4:
		a4 := addr.Sockaddr.(*windows.SockaddrInet4)
		ap = netip.AddrFrom4(a4.Addr)
		port = a4.Port
	case *windows.SockaddrInet6:
		a6 := addr.Sockaddr.(*windows.SockaddrInet6)
		ap = netip.AddrFrom16(a6.Addr)
		port = a6.Port
	default:
		return "unknown sockaddr"
	}
	return netip.AddrPortFrom(ap, uint16(port)).String()
}
