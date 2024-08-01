package fasttool

import (
	"io"
	"net"
	"time"
)

func NewFakeConn(rwc io.ReadWriteCloser) net.Conn {
	return &fakeConn{ReadWriteCloser: rwc}
}

type fakeConn struct {
	io.ReadWriteCloser
}

func (fc *fakeConn) LocalAddr() net.Addr {
	return nil
}

func (fc *fakeConn) RemoteAddr() net.Addr {
	return nil
}

func (fc *fakeConn) SetDeadline(t time.Time) error {
	return nil
}

func (fc *fakeConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (fc *fakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}
