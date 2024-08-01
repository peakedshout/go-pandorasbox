package speedconn

import "net"

type Conn struct {
	net.Conn
	*NetworkSpeedTicker
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		Conn:               conn,
		NetworkSpeedTicker: NewNetworkSpeedTicker(conn),
	}
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return c.NetworkSpeedTicker.upload(b)
}

func (c *Conn) Read(b []byte) (n int, err error) {
	return c.NetworkSpeedTicker.download(b)
}

type PacketConn struct {
	net.PacketConn
	*NetworkSpeedTicker
}

func NewPacketConn(pconn net.PacketConn) *PacketConn {
	return &PacketConn{
		PacketConn:         pconn,
		NetworkSpeedTicker: NewNetworkSpeedTicker(nil),
	}
}

func (pc *PacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	n, err = pc.PacketConn.WriteTo(p, addr)
	if err == nil {
		pc.NetworkSpeedTicker.transfer(n, 0)
	}
	return n, err
}

func (pc *PacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, addr, err = pc.PacketConn.ReadFrom(p)
	if err == nil {
		pc.NetworkSpeedTicker.transfer(0, n)
	}
	return n, addr, err
}
