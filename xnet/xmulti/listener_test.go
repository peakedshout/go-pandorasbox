package xmulti

import (
	"net"
	"testing"
)

func TestNewMultiListener(t *testing.T) {
	l1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	l2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	l, _ := NewMultiListener(nil, true, l1, l2)
	i := 0
	defer func() {
		if i != 2 {
			t.Fatal(i)
		}
	}()
	defer l.Close()
	go func() {
		for {
			_, err := l.Accept()
			if err != nil {
				return
			}
			i++
		}
	}()
	c1, err := net.Dial(l1.Addr().Network(), l1.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()
	c2, err := net.Dial(l1.Addr().Network(), l1.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()
}
