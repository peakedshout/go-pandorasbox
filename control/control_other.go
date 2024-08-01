//go:build !linux && !windows && !darwin

package control

import (
	"errors"
	"strings"
	"syscall"
)

type Sockaddr interface {
	String() string
}

type sockaddr struct {
}

func toSockaddr(addr any) *sockaddr {
	return &sockaddr{}
}

func (s *sockaddr) String() string {
	return addrToString(s)
}

type FdPtr = int

func portReuseControl(network string, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		portReuse(FdPtr(fd))
	})
}

func portReuse(fd FdPtr) error {
	return errors.New("not supported")
}

func rawSockSendTo(fd *FdPtr, b []byte, network, raddr string) (err error) { //nolint:nonamedreturns
	return errors.New("not supported")
}

func rawSockRecvFrom(fd *FdPtr, b []byte, network string) (n int, raddr Sockaddr, err error) {
	return 0, nil, errors.New("not supported")
}

func newSocket(domain, typ, proto int) (FdPtr, error) {
	return 0, errors.New("not supported")
}

func bind(fd FdPtr, addr Sockaddr) error {
	return errors.New("not supported")
}

func toClose(fd FdPtr) error {
	return errors.New("not supported")
}

func recvFrom(fd FdPtr, b []byte) (n int, raddr Sockaddr, err error) {
	return 0, nil, errors.New("not supported")
}

func sendTo(fd FdPtr, b []byte, raddr Sockaddr) error {
	return errors.New("not supported")
}

func send(fd FdPtr, b []byte) error {
	return errors.New("not supported")
}

func write(fd FdPtr, b []byte) (n int, err error) {
	return 0, errors.New("not supported")
}

func read(fd FdPtr, b []byte) (n int, err error) {
	return 0, errors.New("not supported")
}

func connect(fd FdPtr, addr Sockaddr) error {
	return errors.New("not supported")
}

func listen(fd FdPtr, n int) error {
	return errors.New("not supported")
}

func accept(fd FdPtr) (FdPtr, Sockaddr, error) {
	return 0, nil, errors.New("not supported")
}

func parseAddr(v6 bool, addr string) (Sockaddr, bool, error) {
	return nil, false, errors.New("not supported")
}

func parseNetworkToIP(network string) bool {
	if network == "tcp6" || network == "udp6" || strings.HasPrefix(network, "ip6") {
		return true
	} else {
		return false
	}
}

func parseNetworkToType(network string) int {
	return -1
}

func parseIPToDomain(v6 bool) int {
	return -1
}

func parseIPToProto(v6 bool) int {
	return -1
}

func parseNetworkToProto(network string) int {
	return -1
}

func addrToString(addr *sockaddr) string {
	return ""
}
