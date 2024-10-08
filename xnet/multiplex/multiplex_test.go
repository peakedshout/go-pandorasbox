package multiplex

import (
	"bytes"
	"context"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/peakedshout/go-pandorasbox/xnet/fasttool"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMultiplexIO(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	mp := NewMultiplex(context.Background(), 0, 0, 0)
	defer mp.Stop()
	go mp.Listen(func(ctx context.Context) (io.ReadWriteCloser, error) {
		return listen.Accept()
	})
	ctx, cl := context.WithCancel(context.Background())
	defer cl()
	go func() {
		for {
			conn, err := mp.Accept()
			if err != nil {
				return
			}
			go fasttool.EchoConnHandler(ctx, fasttool.NewFakeConn(conn))
		}
	}()
	time.Sleep(1 * time.Second)
	mpd := NewMultiplex(context.Background(), 0, 0, 0)
	defer mpd.Stop()
	xc, err := mpd.Dial(ctx, func(ctx context.Context) (io.ReadWriteCloser, error) {
		return net.Dial(listen.Addr().Network(), listen.Addr().String())
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer xc.Close()
	for i := 0; i < 3; i++ {
		buf := []byte(uuid.NewIdn(4096))
		_, err := xc.Write(buf)
		if err != nil {
			t.Fatal(err)
		}
		newB := make([]byte, len(buf))
		_, err = io.ReadFull(xc, newB)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, newB) {
			t.Fatal("test io failed")
		}
	}
}

func TestMultiplexIO1000(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	mp := NewMultiplex(context.Background(), 0, 0, 0)
	defer mp.Stop()
	go mp.Listen(func(ctx context.Context) (io.ReadWriteCloser, error) {
		return listen.Accept()
	})
	ctx, cl := context.WithCancel(context.Background())
	defer cl()
	go func() {
		for {
			conn, err := mp.Accept()
			if err != nil {
				return
			}
			go fasttool.EchoConnHandler(ctx, fasttool.NewFakeConn(conn))
		}
	}()
	time.Sleep(1 * time.Second)
	mpd := NewMultiplex(context.Background(), 0, 0, 0)
	defer mpd.Stop()
	var wg sync.WaitGroup
	k := int64(0)
	data := uuid.NewIdn(1024 * 32)
	for j := 0; j < 1000; j++ {
		wg.Add(1)
		xc, err := mpd.Dial(ctx, func(ctx context.Context) (io.ReadWriteCloser, error) {
			return net.Dial(listen.Addr().Network(), listen.Addr().String())
		}, 0)
		if err != nil {
			t.Fatal(k, err)
		}
		go func() {
			defer func() {
				fmt.Println(atomic.AddInt64(&k, 1))
			}()
			defer wg.Done()
			defer xc.Close()
			for i := 0; i < 3; i++ {
				buf := []byte(data)
				_, err := xc.Write(buf)
				if err != nil {
					t.Fatal(err)
				}
				newB := make([]byte, len(buf))
				_, err = io.ReadFull(xc, newB)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(buf, newB) {
					t.Fatal("test io failed")
				}
			}
		}()
	}
	wg.Wait()
}

func TestMultiplexIOIdle(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	mp := NewMultiplex(context.Background(), 0, 0, 0)
	defer mp.Stop()
	go mp.Listen(func(ctx context.Context) (io.ReadWriteCloser, error) {
		return listen.Accept()
	})
	ctx, cl := context.WithCancel(context.Background())
	defer cl()
	go func() {
		for {
			conn, err := mp.Accept()
			if err != nil {
				return
			}
			go fasttool.EchoConnHandler(ctx, fasttool.NewFakeConn(conn))
		}
	}()
	time.Sleep(1 * time.Second)
	mpd := NewMultiplex(context.Background(), 0, 0, 0)
	defer mpd.Stop()
	go mpd.DialIDle(func(ctx context.Context) (io.ReadWriteCloser, error) {
		return net.Dial(listen.Addr().Network(), listen.Addr().String())
	}, 3, 0)
	time.Sleep(1 * time.Second)
	xc, err := mpd.Dial(ctx, func(ctx context.Context) (io.ReadWriteCloser, error) {
		return net.Dial(listen.Addr().Network(), listen.Addr().String())
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer xc.Close()
	for i := 0; i < 3; i++ {
		buf := []byte(uuid.NewIdn(4096))
		_, err := xc.Write(buf)
		if err != nil {
			t.Fatal(err)
		}
		newB := make([]byte, len(buf))
		_, err = io.ReadFull(xc, newB)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, newB) {
			t.Fatal("test io failed")
		}
	}
}

func TestMultiplexIO1000Idle(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	mp := NewMultiplex(context.Background(), 0, 0, 0)
	defer mp.Stop()
	go mp.Listen(func(ctx context.Context) (io.ReadWriteCloser, error) {
		return listen.Accept()
	})
	ctx, cl := context.WithCancel(context.Background())
	defer cl()
	go func() {
		for {
			conn, err := mp.Accept()
			if err != nil {
				return
			}
			go fasttool.EchoConnHandler(ctx, fasttool.NewFakeConn(conn))
		}
	}()
	time.Sleep(1 * time.Second)
	mpd := NewMultiplex(context.Background(), 0, 0, 0)
	defer mpd.Stop()
	go mpd.DialIDle(func(ctx context.Context) (io.ReadWriteCloser, error) {
		return net.Dial(listen.Addr().Network(), listen.Addr().String())
	}, 1000, 0)
	time.Sleep(1 * time.Second)
	var wg sync.WaitGroup
	k := int64(0)
	data := uuid.NewIdn(1024 * 32)
	for j := 0; j < 1000; j++ {
		wg.Add(1)
		xc, err := mpd.Dial(ctx, func(ctx context.Context) (io.ReadWriteCloser, error) {
			return net.Dial(listen.Addr().Network(), listen.Addr().String())
		}, 0)
		if err != nil {
			t.Fatal(k, err)
		}
		go func() {
			defer func() {
				fmt.Println(atomic.AddInt64(&k, 1))
			}()
			defer wg.Done()
			defer xc.Close()
			for i := 0; i < 3; i++ {
				buf := []byte(data)
				_, err := xc.Write(buf)
				if err != nil {
					t.Fatal(err)
				}
				newB := make([]byte, len(buf))
				_, err = io.ReadFull(xc, newB)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(buf, newB) {
					t.Fatal("test io failed")
				}
			}
		}()
	}
	wg.Wait()
}
