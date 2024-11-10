package rwcconn

import (
	"io"
	"net"
	"time"
)

func NewRWCConn(rwc io.ReadWriteCloser, laddr, radder net.Addr) net.Conn {
	return &Conn{
		rwc:   rwc,
		laddr: laddr,
		raddr: radder,
	}
}

type Conn struct {
	rwc          io.ReadWriteCloser
	laddr, raddr net.Addr
}

func (c *Conn) Read(b []byte) (n int, err error) {
	return c.rwc.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return c.rwc.Write(b)
}

func (c *Conn) Close() error {
	return c.rwc.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	if i, ok := c.rwc.(interface{ LocalAddr() net.Addr }); ok {
		return i.LocalAddr()
	}
	return c.laddr
}

func (c *Conn) RemoteAddr() net.Addr {
	if i, ok := c.rwc.(interface{ RemoteAddr() net.Addr }); ok {
		return i.RemoteAddr()
	}
	return c.raddr
}

func (c *Conn) SetDeadline(t time.Time) error {
	if i, ok := c.rwc.(interface{ SetDeadline(t time.Time) error }); ok {
		return i.SetDeadline(t)
	}
	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	if i, ok := c.rwc.(interface{ SetReadDeadline(t time.Time) error }); ok {
		return i.SetReadDeadline(t)
	}
	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	if i, ok := c.rwc.(interface{ SetWriteDeadline(t time.Time) error }); ok {
		return i.SetWriteDeadline(t)
	}
	return nil
}
