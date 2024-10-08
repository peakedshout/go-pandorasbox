package xrpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/peakedshout/go-pandorasbox/tool/xpprof"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	_ = server.Serve(listen)
}

func TestNewClientRpc(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	server.AddRpcHandler("test", func(ctx Rpc) (any, error) {
		var str string
		err := ctx.Bind(&str)
		if err != nil {
			return nil, err
		}
		return str, nil
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{}
	client := NewClient(cc)
	defer client.Close()

	sess, err := client.DialContext(ctx, new(net.Dialer), listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	for i := 0; i < 10000; i++ {
		str := uuid.NewIdn(4096)
		var str2 string
		err = sess.Rpc(ctx, "test", str, &str2)
		if err != nil {
			t.Fatal(err)
		}
		if str != str2 {
			t.Fatal()
		}
	}
}

func TestNewClientStream(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	server.AddStreamHandler("test", func(ctx Stream) error {
		for {
			var str string
			err := ctx.Recv(&str)
			if err != nil {
				return err
			}
			err = ctx.Send(str)
			if err != nil {
				return err
			}
		}
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{}
	client := NewClient(cc)
	defer client.Close()

	sess, err := client.DialContext(ctx, new(net.Dialer), listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	ss, err := sess.Stream(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	for i := 0; i < 1000; i++ {
		str := uuid.NewIdn(4096)
		var str2 string
		err = ss.Send(str)
		if err != nil {
			t.Fatal(err)
		}
		err = ss.Recv(&str2)
		if err != nil {
			t.Fatal(err)
		}
		if str != str2 {
			t.Fatal()
		}
	}
	_ = ss.Close()
	if ss.Send("dwdw") == nil {
		t.Fatal()
	}
	if ss.Send(new(string)) == nil {
		t.Fatal()
	}
}

func TestSendClientStream(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	var str string
	ch := make(chan struct{})
	server.AddRecvStreamHandler("test", func(ctx RecvStream) (any, error) {
		for i := 0; i < 10000; i++ {
			err := ctx.Recv(&str)
			if err != nil {
				return nil, err
			}
			ch <- struct{}{}
		}
		str = uuid.NewIdn(4096)
		return str, nil
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{}
	client := NewClient(cc)
	defer client.Close()

	sess, err := client.DialContext(ctx, new(net.Dialer), listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	stream, err := sess.SendStream(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	for i := 0; i < 10000; i++ {
		tmp := uuid.NewIdn(4096)
		err = stream.Send(tmp)
		if err != nil {
			t.Fatal(err)
		}
		<-ch
		if tmp != str {
			t.Fatal()
		}
	}
	time.Sleep(1 * time.Second)
	err = stream.Send("dwdwdw")
	if err == nil {
		t.Fatal()
	}
	var tmp string
	err = stream.Bind(&tmp)
	if err != nil {
		t.Fatal(err)
	}
	if tmp != str {
		t.Fatal()
	}
}

func TestRecvClientStream(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	str := uuid.NewIdn(4096)
	ch := make(chan struct{})
	server.AddSendStreamHandler("test", func(ctx SendStream) error {
		var tmp string
		err := ctx.Bind(&tmp)
		if err != nil {
			return err
		}
		if str != tmp {
			return errors.New("failed")
		}
		for i := 0; i < 10000; i++ {
			str = uuid.NewIdn(4096)
			err = ctx.Send(str)
			if err != nil {
				return err
			}
			ch <- struct{}{}
		}
		return nil
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{}
	client := NewClient(cc)
	defer client.Close()

	sess, err := client.DialContext(ctx, new(net.Dialer), listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	stream, err := sess.RecvStream(ctx, "test", str)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	for i := 0; i < 10000; i++ {
		var tmp string
		err = stream.Recv(&tmp)
		if err != nil {
			t.Fatal(err)
		}
		if tmp != str {
			t.Fatal()
		}
		<-ch
	}
}

func TestNewClientRpc2(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	server.AddRpcHandler("test", func(ctx Rpc) (any, error) {
		var str string
		err := ctx.Bind(&str)
		if err != nil {
			return nil, err
		}
		return str, nil
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{}
	client := NewClient(cc)
	defer client.Close()

	conn, err := net.Dial(listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	sess, err := client.WithConn(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	for i := 0; i < 10000; i++ {
		str := uuid.NewIdn(4096)
		var str2 string
		err = sess.Rpc(ctx, "test", str, &str2)
		if err != nil {
			t.Fatal(err)
		}
		if str != str2 {
			t.Fatal()
		}
	}
}

func TestNewClientReverseRpc(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	var wg sync.WaitGroup
	wg.Add(1)
	server.AddReverseRpcHandler("test", func(ctx ReverseRpc) error {
		defer wg.Done()
		var h string
		err := ctx.Bind(&h)
		if err != nil {
			return err
		}
		for i := 0; i < 10000; i++ {
			data := uuid.NewIdn(4096)
			var resp string
			err = ctx.Rpc(ctx.Context(), h, data, &resp)
			if err != nil {
				return err
			}
			if data != resp {
				t.Error()
				return errors.New("failed")
			}
		}
		return nil
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{}
	client := NewClient(cc)
	defer client.Close()

	sess, err := client.DialContext(ctx, new(net.Dialer), listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	go sess.ReverseRpc(ctx, "test", "subTest", map[string]ClientReverseRpcHandler{
		"subTest": func(rpcContext ClientReverseRpcContext) (any, error) {
			var str string
			err := rpcContext.Bind(&str)
			if err != nil {
				return nil, err
			}
			return str, nil
		},
	})
	wg.Wait()
}

func TestNewClientReverseRpc2(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	var xxx ReverseRpc
	server.AddReverseRpcHandler("test", func(ctx ReverseRpc) error {
		var h string
		err := ctx.Bind(&h)
		if err != nil {
			return err
		}
		xxx = ctx
		<-ctx.Context().Done()
		return nil
	})
	server.AddRpcHandler("test", func(ctx Rpc) (any, error) {
		var str string
		err := ctx.Bind(&str)
		if err != nil {
			return nil, err
		}
		err = xxx.Rpc(ctx.Context(), "subTest", str, &str)
		return str, err
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{}
	client := NewClient(cc)
	defer client.Close()

	sess, err := client.DialContext(ctx, new(net.Dialer), listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	go sess.ReverseRpc(ctx, "test", "subTest", map[string]ClientReverseRpcHandler{
		"subTest": func(rpcContext ClientReverseRpcContext) (any, error) {
			var str string
			err := rpcContext.Bind(&str)
			if err != nil {
				return nil, err
			}
			return str, nil
		},
	})
	time.Sleep(1 * time.Second)
	for i := 0; i < 10000; i++ {
		str := uuid.NewIdn(4096)
		var str2 string
		err = sess.Rpc(ctx, "test", str, &str2)
		if err != nil {
			t.Fatal(err)
		}
		if str != str2 {
			t.Fatal()
		}
	}
}

func TestNewClientReverseRpcx(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	var xxx Stream
	server.AddStreamHandler("test1", func(ctx Stream) error {
		xxx = ctx
		<-ctx.Context().Done()
		return nil
	})
	server.AddRpcHandler("test", func(ctx Rpc) (any, error) {
		var str string
		err := ctx.Bind(&str)
		if err != nil {
			return nil, err
		}
		err = xxx.Send(str)
		if err != nil {
			return nil, err
		}
		err = xxx.Recv(&str)
		if err != nil {
			return nil, err
		}
		return str, nil
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{}
	client := NewClient(cc)
	defer client.Close()

	sess, err := client.DialContext(ctx, new(net.Dialer), listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	go func() {
		stream, err := sess.Stream(ctx, "test1")
		if err != nil {
			t.Error(err)
			return
		}
		_ = stream.Send(nil)
		defer stream.Close()
		for {
			var str string
			err = stream.Recv(&str)
			if err != nil {
				return
			}
			err = stream.Send(str)
			if err != nil {
				return
			}
		}
	}()

	time.Sleep(1 * time.Second)
	for i := 0; i < 10000; i++ {
		str := uuid.NewIdn(4096)
		var str2 string
		err = sess.Rpc(ctx, "test", str, &str2)
		if err != nil {
			t.Fatal(err)
		}
		if str != str2 {
			t.Fatal()
		}
	}
}

func TestSpeed(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 20*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	server.AddStreamHandler("test", func(ctx Stream) error {
		go func() {
			_ = ctxtool.RunTimerFunc(ctx.Context(), 1*time.Second, func(_ context.Context) error {
				fmt.Println("server stream:", fmt.Sprint(ctx.SpeedView()))
				return nil
			})
		}()
		for {
			var str string
			err := ctx.Recv(&str)
			if err != nil {
				return err
			}
			err = ctx.Send(str)
			if err != nil {
				return err
			}
		}
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{}
	client := NewClient(cc)
	defer client.Close()

	sess, err := client.DialContext(ctx, new(net.Dialer), listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	ss, err := sess.Stream(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	go func() {
		_ = ctxtool.RunTimerFunc(ctx, 1*time.Second, func(ctx context.Context) error {
			fmt.Println("client session:", fmt.Sprint(sess.SpeedView()))
			fmt.Println("client stream:", fmt.Sprint(ss.SpeedView()))
			return nil
		})
	}()
	str := uuid.NewIdn(4096)
	for i := 0; i < 200000; i++ {
		var str2 string
		err = ss.Send(str)
		if err != nil {
			t.Fatal(err)
		}
		err = ss.Recv(&str2)
		if err != nil {
			t.Fatal(err)
		}
		if str != str2 {
			t.Fatal()
		}
	}
	_ = ss.Close()
	if ss.Send("dwdw") == nil {
		t.Fatal()
	}
	if ss.Send(new(string)) == nil {
		t.Fatal()
	}
	time.Sleep(2 * time.Second)
}

func TestSessMap(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 20*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx, CacheTime: 2 * time.Second}
	server := NewServer(sc)
	defer server.Close()
	server.AddStreamHandler("test", func(ctx Stream) error {
		for {
			var str string
			err := ctx.Recv(&str)
			if err != nil {
				return err
			}
			err = ctx.Send(str)
			if err != nil {
				return err
			}
		}
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{CacheTime: 2 * time.Second}
	client := NewClient(cc)
	defer client.Close()

	sl := make([]*ClientSession, 0, 1000)
	for i := 0; i < 1000; i++ {
		sess, err := client.DialContext(ctx, new(net.Dialer), listen.Addr().Network(), listen.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		sl = append(sl, sess)
	}
	for i := 0; i < 1000; i++ {
		_ = sl[i].Close()
	}
	_ = client.Close()
}

func testMem(ctx context.Context) func() {
	xpprof.AsyncListenAddress("127.0.0.1:7778")
	var m runtime.MemStats
	go ctxtool.RunTimerFunc(ctx, 1*time.Second, func(ctx context.Context) error {
		runtime.ReadMemStats(&m)
		fmt.Printf("G = %v ", runtime.NumGoroutine())
		fmt.Printf("\tAlloc = %0.2f MiB", float64(m.Alloc)/1024/1024)
		fmt.Printf("\tTotalAlloc = %0.2f MiB", float64(m.TotalAlloc)/1024/1024)
		fmt.Printf("\tSys = %0.2f MiB", float64(m.Sys)/1024/1024)
		fmt.Printf("\tNumGC = %v\n", m.NumGC)
		runtime.GC()
		return nil
	})
	return func() {
		<-ctx.Done()
	}
}

func TestStreamMem(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 60*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	server.AddStreamHandler("test", func(ctx Stream) error {
		for {
			var str string
			err := ctx.Recv(&str)
			if err != nil {
				return err
			}
			err = ctx.Send(str)
			if err != nil {
				return err
			}
		}
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)

	defer testMem(ctx)()

	cc := &ClientConfig{ShareDialFunc: func(ctx context.Context) (net.Conn, error) {
		return new(net.Dialer).DialContext(ctx, listen.Addr().Network(), listen.Addr().String())
	}}
	client := NewClient(cc)
	defer client.Close()

	str := uuid.NewIdn(4096)

	sess, err := client.DialContext(ctx, new(net.Dialer), listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	time.Sleep(2 * time.Second)

	var wg sync.WaitGroup
	for j := 0; j < 2000; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testStream(str, sess, t)
		}()
	}
	wg.Wait()
}

func testStream(str string, sess *ClientSession, t *testing.T) {
	ss, err := sess.c.Stream(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	for i := 0; i < 100; i++ {
		var str2 string
		err = ss.Send(str)
		if err != nil {
			t.Fatal(err)
		}
		err = ss.Recv(&str2)
		if err != nil {
			t.Fatal(err)
		}
		if str != str2 {
			t.Fatal()
		}
	}
}

func TestClientShareRpc(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	server.AddRpcHandler("test", func(ctx Rpc) (any, error) {
		var str string
		err := ctx.Bind(&str)
		if err != nil {
			return nil, err
		}
		return str, nil
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{ShareDialFunc: func(ctx context.Context) (net.Conn, error) {
		return new(net.Dialer).DialContext(ctx, listen.Addr().Network(), listen.Addr().String())
	}}
	client := NewClient(cc)
	defer client.Close()

	for i := 0; i < 10000; i++ {
		str := uuid.NewIdn(4096)
		var str2 string
		err = client.Rpc(ctx, "test", str, &str2)
		if err != nil {
			t.Fatal(err)
		}
		if str != str2 {
			t.Fatal()
		}
	}
}

func TestClientShareStream(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	server.AddStreamHandler("test", func(ctx Stream) error {
		for {
			var str string
			err := ctx.Recv(&str)
			if err != nil {
				return err
			}
			err = ctx.Send(str)
			if err != nil {
				return err
			}
		}
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{ShareDialFunc: func(ctx context.Context) (net.Conn, error) {
		return new(net.Dialer).DialContext(ctx, listen.Addr().Network(), listen.Addr().String())
	}}
	client := NewClient(cc)
	defer client.Close()

	ss, err := client.Stream(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	for i := 0; i < 1000; i++ {
		str := uuid.NewIdn(4096)
		var str2 string
		err = ss.Send(str)
		if err != nil {
			t.Fatal(err)
		}
		err = ss.Recv(&str2)
		if err != nil {
			t.Fatal(err)
		}
		if str != str2 {
			t.Fatal()
		}
	}
	_ = ss.Close()
	if ss.Send("dwdw") == nil {
		t.Fatal()
	}
	if ss.Send(new(string)) == nil {
		t.Fatal()
	}
}

func TestClientShareSendStream(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	var str string
	ch := make(chan struct{})
	server.AddRecvStreamHandler("test", func(ctx RecvStream) (any, error) {
		for i := 0; i < 10000; i++ {
			err := ctx.Recv(&str)
			if err != nil {
				return nil, err
			}
			ch <- struct{}{}
		}
		str = uuid.NewIdn(4096)
		return str, nil
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{ShareDialFunc: func(ctx context.Context) (net.Conn, error) {
		return new(net.Dialer).DialContext(ctx, listen.Addr().Network(), listen.Addr().String())
	}}
	client := NewClient(cc)
	defer client.Close()

	stream, err := client.SendStream(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	for i := 0; i < 10000; i++ {
		tmp := uuid.NewIdn(4096)
		err = stream.Send(tmp)
		if err != nil {
			t.Fatal(err)
		}
		<-ch
		if tmp != str {
			t.Fatal()
		}
	}
	time.Sleep(1 * time.Second)
	err = stream.Send("dwdwdw")
	if err == nil {
		t.Fatal()
	}
	var tmp string
	err = stream.Bind(&tmp)
	if err != nil {
		t.Fatal(err)
	}
	if tmp != str {
		t.Fatal()
	}
}

func TestClientShareRecvStream(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	str := uuid.NewIdn(4096)
	ch := make(chan struct{})
	server.AddSendStreamHandler("test", func(ctx SendStream) error {
		var tmp string
		err := ctx.Bind(&tmp)
		if err != nil {
			return err
		}
		if str != tmp {
			return errors.New("failed")
		}
		for i := 0; i < 10000; i++ {
			str = uuid.NewIdn(4096)
			err = ctx.Send(str)
			if err != nil {
				return err
			}
			ch <- struct{}{}
		}
		return nil
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{ShareDialFunc: func(ctx context.Context) (net.Conn, error) {
		return new(net.Dialer).DialContext(ctx, listen.Addr().Network(), listen.Addr().String())
	}}
	client := NewClient(cc)
	defer client.Close()

	stream, err := client.RecvStream(ctx, "test", str)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	for i := 0; i < 10000; i++ {
		var tmp string
		err = stream.Recv(&tmp)
		if err != nil {
			t.Fatal(err)
		}
		if tmp != str {
			t.Fatal()
		}
		<-ch
	}
}

func TestClientShareReverseRpc(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	sc := &ServerConfig{Ctx: ctx}
	server := NewServer(sc)
	defer server.Close()
	var wg sync.WaitGroup
	wg.Add(1)
	server.AddReverseRpcHandler("test", func(ctx ReverseRpc) error {
		defer wg.Done()
		var h string
		err := ctx.Bind(&h)
		if err != nil {
			return err
		}
		for i := 0; i < 10000; i++ {
			data := uuid.NewIdn(4096)
			var resp string
			err = ctx.Rpc(ctx.Context(), h, data, &resp)
			if err != nil {
				return err
			}
			if data != resp {
				t.Error()
				return errors.New("failed")
			}
		}
		return nil
	})
	listen, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)
	cc := &ClientConfig{ShareDialFunc: func(ctx context.Context) (net.Conn, error) {
		return new(net.Dialer).DialContext(ctx, listen.Addr().Network(), listen.Addr().String())
	}}
	client := NewClient(cc)
	defer client.Close()

	go client.ReverseRpc(ctx, "test", "subTest", map[string]ClientReverseRpcHandler{
		"subTest": func(rpcContext ClientReverseRpcContext) (any, error) {
			var str string
			err := rpcContext.Bind(&str)
			if err != nil {
				return nil, err
			}
			return str, nil
		},
	})
	wg.Wait()
}
