package xnetutil

import (
	"net"
	"testing"
)

func TestNewUpgraderListener(t *testing.T) {

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	listener := NewUpgraderListener(ln, EmptyUpgrader(), 0, 0)
	go func() {
		_, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
	}()
	_, err = net.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
}
