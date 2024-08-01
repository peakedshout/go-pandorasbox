package xdummy

import (
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"io"
	"net"
	"time"
)

func NewDummyConn() (net.Conn, net.Conn) {
	p1r, p1w := io.Pipe()
	p2r, p2w := io.Pipe()
	conn1 := &dummyConn{
		read:  p1r,
		write: p2w,
	}
	conn2 := &dummyConn{
		read:  p2r,
		write: p1w,
	}
	return conn1, conn2
}

type dummyConn struct {
	read  *io.PipeReader
	write *io.PipeWriter
}

func (c *dummyConn) Read(b []byte) (n int, err error) {
	return c.read.Read(b)
}

func (c *dummyConn) Write(b []byte) (n int, err error) {
	return c.write.Write(b)
}

func (c *dummyConn) Close() error {
	_ = c.write.Close()
	_ = c.read.Close()
	return nil
}

func (c *dummyConn) LocalAddr() net.Addr {
	return xnetutil.NewNetAddr("dummy", "dummy")
}

func (c *dummyConn) RemoteAddr() net.Addr {
	return xnetutil.NewNetAddr("dummy", "dummy")
}

func (c *dummyConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *dummyConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *dummyConn) SetWriteDeadline(t time.Time) error {
	return nil
}
