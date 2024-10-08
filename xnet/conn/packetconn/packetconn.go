package packetconn

import (
	"net"
	"time"
)

type packetConn struct {
	net.PacketConn
	raddr net.Addr
}

func NewPacketConn(conn net.PacketConn, raddr net.Addr) net.Conn {
	return &packetConn{
		PacketConn: conn,
		raddr:      raddr,
	}
}

func (p *packetConn) Read(b []byte) (n int, err error) {
	for {
		n, addr, err := p.PacketConn.ReadFrom(b)
		if err != nil {
			return 0, err
		}
		if addr != p.raddr {
			continue
		}
		return n, nil
	}
}

func (p *packetConn) Write(b []byte) (n int, err error) {
	return p.PacketConn.WriteTo(b, p.raddr)
}

func (p *packetConn) Close() error {
	return p.PacketConn.Close()
}

func (p *packetConn) LocalAddr() net.Addr {
	return p.PacketConn.LocalAddr()
}

func (p *packetConn) RemoteAddr() net.Addr {
	return p.raddr
}

func (p *packetConn) SetDeadline(t time.Time) error {
	return p.PacketConn.SetDeadline(t)
}

func (p *packetConn) SetReadDeadline(t time.Time) error {
	return p.PacketConn.SetReadDeadline(t)
}

func (p *packetConn) SetWriteDeadline(t time.Time) error {
	return p.SetWriteDeadline(t)
}
