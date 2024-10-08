package fasttool

import (
	"context"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"net"
	"net/http"
)

func UdpPacketListener(handler PacketHandler, addr ...string) (net.PacketConn, error) {
	return UdpPacketListenerContext(context.Background(), handler, addr...)
}

func UdpPacketListenerContext(ctx context.Context, handler PacketHandler, addr ...string) (net.PacketConn, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	laddr := "127.0.0.1:"
	if len(addr) != 0 {
		laddr = addr[0]
	}

	lc := net.ListenConfig{}
	packetConn, err := lc.ListenPacket(ctx, "udp", laddr)
	if err != nil {
		return nil, err
	}
	nctx, cl := context.WithCancel(ctx)

	go func() {
		go ctxtool.WaitFunc(nctx, func() {
			_ = packetConn.Close()
		})
		defer cl()
		buf := make([]byte, 32*1024)
		for {
			n, addr, err := packetConn.ReadFrom(buf)
			if err != nil {
				return
			}
			data := make([]byte, n)
			copy(data, buf[:n])
			err = handler(packetConn, data, addr)
			if err != nil {
				return
			}
		}
	}()

	return packetConn, nil
}

func TcpListener(handler ConnHandler, addr ...string) (net.Listener, error) {
	return TcpListenerContext(context.Background(), handler, addr...)
}

func TcpListenerContext(ctx context.Context, handler ConnHandler, addr ...string) (net.Listener, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	laddr := "127.0.0.1:"
	if len(addr) != 0 {
		laddr = addr[0]
	}

	lc := net.ListenConfig{}
	listen, err := lc.Listen(ctx, "tcp", laddr)
	if err != nil {
		return nil, err
	}

	nctx, cl := context.WithCancel(ctx)

	go func() {
		go ctxtool.WaitFunc(nctx, func() {
			_ = listen.Close()
		})
		defer cl()
		for {
			conn, err := listen.Accept()
			if err != nil {
				return
			}
			go handler(nctx, conn)
		}
	}()

	return listen, nil
}

func Http(handler http.Handler, addr ...string) (*http.Server, string, error) {
	return HttpContext(context.Background(), handler, addr...)
}

func HttpContext(ctx context.Context, handler http.Handler, addr ...string) (*http.Server, string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	laddr := "127.0.0.1:"
	if len(addr) != 0 {
		laddr = addr[0]
	}

	lc := net.ListenConfig{}
	listen, err := lc.Listen(ctx, "tcp", laddr)
	if err != nil {
		return nil, "", err
	}

	server := &http.Server{Handler: handler}

	nctx, cl := context.WithCancel(ctx)
	go func() {
		go ctxtool.WaitFunc(nctx, func() {
			_ = server.Close()
		})
		defer cl()
		_ = server.Serve(listen)
	}()

	return server, fmt.Sprintf("http://%s", listen.Addr().String()), nil
}
