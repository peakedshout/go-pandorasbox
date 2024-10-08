package flowconn

import (
	"github.com/peakedshout/go-pandorasbox/tool/bio"
	"net"
)

type Conn struct {
	bio.FlowReadWriter
	net.Conn
}

func (c *Conn) Read(b []byte) (n int, err error) {
	return c.FlowReadWriter.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return c.FlowReadWriter.Write(b)
}

func NewConn(conn net.Conn) net.Conn {
	return &Conn{
		FlowReadWriter: bio.NewFlowReadWriter(conn),
		Conn:           conn,
	}
}
