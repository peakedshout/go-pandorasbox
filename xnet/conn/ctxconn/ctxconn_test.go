package ctxconn

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/xnet/xquic"
	"net"
	"testing"
	"time"
)

func TestNewCtxConn(t *testing.T) {
	conn, err := net.Dial("tcp", "www.baidu.com:443")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	ctxConn := NewCtxConn(ctx, conn)
	_ = conn.Close()
	time.Sleep(1 * time.Second)
	if ctxConn.Err() == nil {
		t.Fatal()
	}

	conn, err = net.Dial("tcp", "www.baidu.com:443")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ctx, cl := context.WithCancel(ctx)
	ctxConn = NewCtxConn(ctx, conn)
	cl()
	time.Sleep(1 * time.Second)
	if ctxConn.Err() == nil {
		t.Fatal()
	}
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err == nil {
		t.Fatal()
	}
}

func TestNewCtxConnQ(t *testing.T) {
	ln, err2 := xquic.Listen("udp", "127.0.0.1:0")
	if err2 != nil {
		t.Fatal(err2)
	}
	defer ln.Close()
	go func() {
		buf := make([]byte, 1024)
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				for {
					_, err = conn.Read(buf)
					if err != nil {
						conn.Close()
						continue
					}
				}
			}()
		}
	}()
	ctx := context.WithValue(context.Background(), xquic.XQuicPreHandShake, struct{}{})
	conn, err := xquic.DialContext(ctx, "udp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	ctxConn := NewCtxConn(ctx, conn)
	_ = conn.Close()
	time.Sleep(2 * time.Second)
	if ctxConn.Err() == nil {
		t.Fatal()
	}
	conn, err = xquic.DialContext(ctx, "udp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ctx, cl := context.WithCancel(ctx)
	ctxConn = NewCtxConn(ctx, conn)
	cl()
	time.Sleep(2 * time.Second)
	if ctxConn.Err() == nil {
		t.Fatal()
	}
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err == nil {
		t.Fatal()
	}
}
