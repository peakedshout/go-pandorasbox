package websocketconn

import (
	"bytes"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"testing"
)

func TestNewConn(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		conn, err := listen.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		wsconn := Server(conn, nil)
		buf := make([]byte, 1024)
		for {
			n, err := wsconn.Read(buf)
			if err != nil {
				return
			}
			_, err = wsconn.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}()
	conn, err := net.Dial(listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	wsconn := Client(conn, nil)
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = wsconn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := wsconn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}

func TestClient(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	surl := "ws://"
	ourl := "http://"

	cfg, _ := websocket.NewConfig(surl+listen.Addr().String()+"/", ourl+listen.Addr().String()+"/")
	go func() {
		_ = http.Serve(listen, websocket.Server{
			Config: *cfg,
			Handler: func(conn *websocket.Conn) {
				buf := make([]byte, 1024)
				for {
					n, err := conn.Read(buf)
					if err != nil {
						return
					}
					_, err = conn.Write(buf[:n])
					if err != nil {
						return
					}
				}
			},
		})
	}()
	conn, err := net.Dial(listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	wsconn := Client(conn, nil)
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = wsconn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := wsconn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}

func TestServer(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		conn, err := listen.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		wsconn := Server(conn, nil)
		buf := make([]byte, 1024)
		for {
			n, err := wsconn.Read(buf)
			if err != nil {
				return
			}
			_, err = wsconn.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}()
	conn, err := net.Dial(listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	surl := "ws://"
	ourl := "http://"
	config, err := websocket.NewConfig(surl+conn.RemoteAddr().String(), ourl+conn.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	wsconn, err := websocket.NewClient(config, conn)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = wsconn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := wsconn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}

func TestNewConnTLS(t *testing.T) {
	tlsConfig, err := pcrypto.NewDefaultTlsConfig()
	if err != nil {
		t.Fatal(err)
	}
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		conn, err := listen.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		wsconn := Server(conn, tlsConfig)
		buf := make([]byte, 1024)
		for {
			n, err := wsconn.Read(buf)
			if err != nil {
				return
			}
			_, err = wsconn.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}()
	conn, err := net.Dial(listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	wsconn := Client(conn, tlsConfig)
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = wsconn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := wsconn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}

func TestClientTLS(t *testing.T) {
	tlsConfig, err := pcrypto.NewDefaultTlsConfig()
	if err != nil {
		t.Fatal(err)
	}
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	listen = tls.NewListener(listen, tlsConfig)

	surl := "wss://"
	ourl := "https://"

	cfg, _ := websocket.NewConfig(surl+listen.Addr().String()+"/", ourl+listen.Addr().String()+"/")
	go func() {
		_ = http.Serve(listen, websocket.Server{
			Config: *cfg,
			Handler: func(conn *websocket.Conn) {
				buf := make([]byte, 1024)
				for {
					n, err := conn.Read(buf)
					if err != nil {
						return
					}
					_, err = conn.Write(buf[:n])
					if err != nil {
						return
					}
				}
			},
		})
	}()
	conn, err := net.Dial(listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	wsconn := Client(conn, tlsConfig)
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = wsconn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := wsconn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}

func TestServerTLS(t *testing.T) {
	tlsConfig, err := pcrypto.NewDefaultTlsConfig()
	if err != nil {
		t.Fatal(err)
	}
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		conn, err := listen.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		wsconn := Server(conn, tlsConfig)
		buf := make([]byte, 1024)
		for {
			n, err := wsconn.Read(buf)
			if err != nil {
				return
			}
			_, err = wsconn.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}()
	conn, err := net.Dial(listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn = tls.Client(conn, tlsConfig)
	surl := "wss://"
	ourl := "https://"
	config, err := websocket.NewConfig(surl+conn.RemoteAddr().String(), ourl+conn.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	wsconn, err := websocket.NewClient(config, conn)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = wsconn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := wsconn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}
