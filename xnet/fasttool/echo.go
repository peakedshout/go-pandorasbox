package fasttool

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"io"
	"net"
	"net/http"
)

func EchoUdpPacketListener(addr ...string) (net.PacketConn, error) {
	return EchoUdpPacketListenerContext(context.Background(), addr...)
}

func EchoUdpPacketListenerContext(ctx context.Context, addr ...string) (net.PacketConn, error) {
	return UdpPacketListenerContext(ctx, EchoPacketHandler, addr...)
}

func EchoTcpListener(addr ...string) (net.Listener, error) {
	return EchoTcpListenerContext(context.Background(), addr...)
}

func EchoTcpListenerContext(ctx context.Context, addr ...string) (net.Listener, error) {
	return TcpListenerContext(ctx, EchoConnHandler, addr...)
}

func EchoHttp(addr ...string) (*http.Server, string, error) {
	return EchoHttpContext(context.Background(), addr...)
}

func EchoHttpContext(ctx context.Context, addr ...string) (*http.Server, string, error) {
	return HttpContext(ctx, EchoHttpHandler, addr...)
}

var EchoConnHandler ConnHandler = func(ctx context.Context, conn net.Conn) error {
	nctx, cl := context.WithCancel(ctx)
	defer cl()
	go ctxtool.WaitFunc(nctx, func() {
		_ = conn.Close()
	})
	_, err := io.Copy(conn, conn)
	return err
}

var EchoPacketHandler PacketHandler = func(pconn net.PacketConn, data []byte, addr net.Addr) error {
	_, err := pconn.WriteTo(data, addr)
	if err != nil {
		return err
	}
	return nil
}

var EchoHttpHandler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
	for k := range writer.Header() {
		writer.Header().Del(k)
	}
	for k, v := range request.Header {
		for _, s := range v {
			writer.Header().Add(k, s)
		}
	}
	body := request.Body
	if body != nil {
		all, err := io.ReadAll(body)
		if err != nil {
			_, _ = writer.Write([]byte(err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
		} else {
			_, err := writer.Write(all)
			if err != nil {
				_, _ = writer.Write([]byte(err.Error()))
				writer.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
}
