package xdummy

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"io"
	"net"
	"sync"
	"testing"
)

func TestNewDummyListenDialer(t *testing.T) {
	ctx := context.Background()
	l, d := NewDummyListenDialer(ctx)
	defer l.Close()
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(conn2 net.Conn) {
				defer conn2.Close()
				_, _ = io.Copy(conn2, conn2)
			}(conn)
		}
	}()
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := d.Dial("", "")
			if err != nil {
				t.Error(err)
				return
			}
			defer conn.Close()
			buf := make([]byte, 4096)
			for j := 0; j < 100; j++ {
				str := uuid.NewIdn(4096)
				_, err := conn.Write([]byte(str))
				if err != nil {
					t.Error(err)
					return
				}
				_, err = io.ReadFull(conn, buf)
				if err != nil {
					t.Error(err)
					return
				}
				if str != string(buf) {
					t.Error("bad")
					return
				}
			}
		}()
	}
	wg.Wait()
}
