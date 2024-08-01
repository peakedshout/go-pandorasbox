package control

import (
	"math/rand"
	"net/netip"
	"syscall"
)

func Close(fd FdPtr) error {
	return toClose(fd)
}

func RawSockSendTcpSYNMessage(fd *FdPtr, b []byte, network, raddr, laddr string) (err error) {
	rap, err := netip.ParseAddrPort(raddr)
	if err != nil {
		return err
	}
	lap, err := netip.ParseAddrPort(laddr)
	if err != nil {
		return err
	}
	theader := newTcpMessage(lap.Port(), rap.Port(), b)
	theader.SrcIp = rap.Addr().AsSlice()
	theader.DstIp = lap.Addr().AsSlice()
	theader.SeqNum = rand.Uint32() << 1
	theader.Window = 0xffff
	theader.SetFlag(SYN, true)
	theader.RawOptions = DefaultRawOptions()
	tb, err := theader.Bytes()
	if err != nil {
		return err
	}
	return RawSockSendTo(fd, tb, network, raddr)
}

func RawSockSendTcpMessage(fd *FdPtr, b []byte, network, raddr, laddr string, fn func(th *TCPHeader)) (err error) {
	rap, err := netip.ParseAddrPort(raddr)
	if err != nil {
		return err
	}
	lap, err := netip.ParseAddrPort(laddr)
	if err != nil {
		return err
	}
	theader := newTcpMessage(lap.Port(), rap.Port(), b)
	theader.SrcIp = rap.Addr().AsSlice()
	theader.DstIp = lap.Addr().AsSlice()
	fn(theader.TCPHeader)
	tb, err := theader.Bytes()
	if err != nil {
		return err
	}
	return RawSockSendTo(fd, tb, network, raddr)
}

func RawSockSendTo(fd *FdPtr, b []byte, network, raddr string) (err error) {
	return rawSockSendTo(fd, b, network, raddr)
}

func RawSockRecvFrom(fd *FdPtr, b []byte, network string) (n int, raddr Sockaddr, err error) {
	return rawSockRecvFrom(fd, b, network)
}

func NewSocket(domain, typ, proto int) (FdPtr, error) {
	return newSocket(domain, typ, proto)
}

func Bind(fd FdPtr, addr Sockaddr) error {
	return bind(fd, addr)
}

func RecvFrom(fd FdPtr, b []byte) (n int, raddr Sockaddr, err error) {
	return recvFrom(fd, b)
}

func SendTo(fd FdPtr, b []byte, raddr Sockaddr) error {
	return sendTo(fd, b, raddr)
}

func Send(fd FdPtr, b []byte) error {
	return send(fd, b)
}

func Write(fd FdPtr, b []byte) (n int, err error) {
	return write(fd, b)
}

func Read(fd FdPtr, b []byte) (n int, err error) {
	return read(fd, b)
}

func Connect(fd FdPtr, addr Sockaddr) error {
	return connect(fd, addr)
}

func Listen(fd FdPtr, n int) error {
	return listen(fd, n)
}

func Accept(fd FdPtr) (FdPtr, Sockaddr, error) {
	return accept(fd)
}

func PortReuseControl(network string, address string, c syscall.RawConn) error {
	return portReuseControl(network, address, c)
}

func PortReuse(fd FdPtr) error {
	return portReuse(fd)
}
