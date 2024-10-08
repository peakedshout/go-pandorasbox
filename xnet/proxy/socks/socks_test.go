package socks

import (
	"context"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/peakedshout/go-pandorasbox/xnet/fasttool"
	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/net/proxy"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func testLPConn(t *testing.T) net.PacketConn {
	listener, err := fasttool.EchoUdpPacketListener()
	if err != nil {
		t.Fatal(err)
	}
	return listener
}

func testListen(t *testing.T) net.Listener {
	listener, err := fasttool.EchoTcpListener()
	if err != nil {
		t.Fatal(err)
	}
	return listener
}

func testConn(t *testing.T, conn net.Conn, data string) {
	_, err := conn.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, len(data))
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		t.Fatal(err)
	}
	if data != string(buf) {
		t.Fatal("test failed")
	}
}

func testPConn(t *testing.T, pconn net.PacketConn, addr net.Addr, data string) {
	_, err := pconn.WriteTo([]byte(data), addr)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 255+len(data))
	n, raddr, err := pconn.ReadFrom(buf)
	if err != nil {
		fmt.Println(addr)
		t.Fatal(err)
	}
	if raddr.String() != addr.String() || data != string(buf[:n]) {
		t.Fatal("test failed")
	}
}

func TestNewServer(t *testing.T) {
	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	go func() {
		time.Sleep(3 * time.Second)
		_ = server.Close()
	}()
	err = server.ListenAndServe("tcp", "127.0.0.1:4440")
	if err != nil {
		return
	}
}

func TestSOCKS5(t *testing.T) {
	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)
	dr, err := proxy.SOCKS5(listen.Addr().Network(), listen.Addr().String(), &proxy.Auth{
		User:     "test",
		Password: "test123",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ln := testListen(t)
	defer ln.Close()
	conn, err := dr.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		testConn(t, conn, uuid.NewIdn(4096))
	}
}

func TestSOCKS5CONNCT(t *testing.T) {
	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)
	dr, err := SOCKS5CONNECTP(listen.Addr().Network(), listen.Addr().String(), &S5AuthPassword{
		User:     "test",
		Password: "test123",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ln := testListen(t)
	defer ln.Close()
	conn, err := dr.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		testConn(t, conn, uuid.NewIdn(4096))
	}
}

func TestSOCKS4CONNECT(t *testing.T) {
	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)
	dr, err := SOCKS4CONNECT(listen.Addr().Network(), listen.Addr().String(), S4UserId{1, 2, 3, 4, 5, 6}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ln := testListen(t)
	defer ln.Close()
	conn, err := dr.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		testConn(t, conn, uuid.NewIdn(4096))
	}
}

func TestSOCKS5BIND(t *testing.T) {
	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)
	dx := net.Dialer{LocalAddr: &net.TCPAddr{
		IP:   net.IP{127, 0, 0, 1},
		Port: rand.Intn(10000) + 10000,
		Zone: "",
	}}
	dr, err := SOCKS5BINDP(listen.Addr().Network(), listen.Addr().String(), &S5AuthPassword{
		User:     "test",
		Password: "test123",
	}, nil, func(addr net.Addr) error {
		conn, err := dx.Dial(addr.Network(), addr.String())
		if err != nil {
			return err
		}
		go func() {
			defer conn.Close()
			io.Copy(conn, conn)
		}()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	conn, err := dr.Dial(dx.LocalAddr.Network(), dx.LocalAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		testConn(t, conn, uuid.NewIdn(4096))
	}
}

func TestSOCKS4BIND(t *testing.T) {
	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)
	dx := net.Dialer{LocalAddr: &net.TCPAddr{
		IP:   net.IP{127, 0, 0, 1},
		Port: rand.Intn(10000) + 10000,
		Zone: "",
	}}
	dr, err := SOCKS4BIND(listen.Addr().Network(), listen.Addr().String(), S4UserId{1, 2, 3, 4, 5, 6}, nil, func(addr net.Addr) error {
		conn, err := dx.Dial(addr.Network(), addr.String())
		if err != nil {
			return err
		}
		go func() {
			defer conn.Close()
			io.Copy(conn, conn)
		}()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	conn, err := dr.Dial(dx.LocalAddr.Network(), dx.LocalAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		testConn(t, conn, uuid.NewIdn(4096))
	}
}

func TestSOCKS5UDPASSOCIATEP(t *testing.T) {
	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)
	pConn := testLPConn(t)
	defer pConn.Close()
	ucfg, err := SOCKS5UDPASSOCIATEP(listen.Addr().Network(), listen.Addr().String(), &S5AuthPassword{
		User:     "test",
		Password: "test123",
	}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	pConn2, err := ucfg.ListenPacket("udp", "0:12133")
	if err != nil {
		t.Fatal(err)
	}
	defer pConn2.Close()
	for i := 0; i < 3; i++ {
		testPConn(t, pConn2, pConn.LocalAddr(), uuid.NewIdn(4096))
	}
}

func TestDNS(t *testing.T) {
	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", ":9978")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go server.Serve(listen)

	time.Sleep(1 * time.Second)

	pconn, err := SOCKS5UDPASSOCIATEP(listen.Addr().Network(), listen.Addr().String(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	xconn, err := pconn.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		t.Fatal(err)
	}
	defer xconn.Close()
	m := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID: 0,
		},
		Questions: []dnsmessage.Question{
			{
				Name:  dnsmessage.MustNewName("www.google.com."),
				Type:  dnsmessage.TypeALL,
				Class: dnsmessage.ClassINET,
			},
		},
	}
	m.ID = 3
	b, err := m.Pack()
	if err != nil {
		t.Fatal(err)
	}
	udpAddr, _ := net.ResolveUDPAddr("", "8.8.8.8:53")
	_, err = xconn.WriteTo(b, udpAddr)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 4096)
	n, _, err := xconn.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}
	var m1 dnsmessage.Message
	err = m1.Unpack(buf[:n])
	fmt.Println(m1.GoString())

	xconn2, err := pconn.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		t.Fatal(err)
	}
	defer xconn2.Close()

	m.ID = 4
	b, err = m.Pack()
	if err != nil {
		t.Fatal(err)
	}
	_, err = xconn2.WriteTo(b, udpAddr)
	if err != nil {
		t.Fatal(err)
	}
	n, _, err = xconn2.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}
	err = m1.Unpack(buf[:n])
	fmt.Println(m1.GoString())
}

func TestHttp(t *testing.T) {
	echoHttp, httpAddr, err := fasttool.EchoHttp()
	if err != nil {
		t.Fatal(echoHttp)
	}
	defer echoHttp.Close()

	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: nil,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)

	parse, err := url.Parse(fmt.Sprintf("socks5://test:test123@%s", listen.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}

	c := http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(parse)}}
	resp, err := c.Get(httpAddr)
	if err != nil || resp.StatusCode != 200 {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
}

func testRelayServer(t *testing.T) net.Listener {
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		defer listen.Close()
		for {
			conn, err := listen.Accept()
			if err != nil {
				return
			}
			go func(conn2 net.Conn) {
				defer conn2.Close()
				_ = RelayServe(conn2)
			}(conn)
		}
	}()
	return listen
}

func TestSOCKS5CONNECTRelay(t *testing.T) {
	relay := testRelayServer(t)
	defer relay.Close()

	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	cfg.CMDConfig.CMDCONNECTHandler = RelayCMDCONNECTHandler(func(ctx context.Context) (net.Conn, error) {
		dr := net.Dialer{}
		return dr.DialContext(ctx, relay.Addr().Network(), relay.Addr().String())
	})
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)
	dr, err := SOCKS5CONNECTP(listen.Addr().Network(), listen.Addr().String(), &S5AuthPassword{
		User:     "test",
		Password: "test123",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ln := testListen(t)
	defer ln.Close()
	conn, err := dr.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		testConn(t, conn, uuid.NewIdn(4096))
	}
}

func TestSOCKS4CONNECTRelay(t *testing.T) {
	relay := testRelayServer(t)
	defer relay.Close()

	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	cfg.CMDConfig.CMDCONNECTHandler = RelayCMDCONNECTHandler(func(ctx context.Context) (net.Conn, error) {
		dr := net.Dialer{}
		return dr.DialContext(ctx, relay.Addr().Network(), relay.Addr().String())
	})
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)
	dr, err := SOCKS4CONNECT(listen.Addr().Network(), listen.Addr().String(), S4UserId{1, 2, 3, 4, 5, 6}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ln := testListen(t)
	defer ln.Close()
	conn, err := dr.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		testConn(t, conn, uuid.NewIdn(4096))
	}
}

func TestSOCKS5BINDRelay(t *testing.T) {
	relay := testRelayServer(t)
	defer relay.Close()

	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	cfg.CMDConfig.CMDBINDHandler = RelayCMDBINDHandler(func(ctx context.Context) (net.Conn, error) {
		dr := net.Dialer{}
		return dr.DialContext(ctx, relay.Addr().Network(), relay.Addr().String())
	})
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)
	dx := net.Dialer{LocalAddr: &net.TCPAddr{
		IP:   net.IP{127, 0, 0, 1},
		Port: rand.Intn(10000) + 10000,
		Zone: "",
	}}
	dr, err := SOCKS5BINDP(listen.Addr().Network(), listen.Addr().String(), &S5AuthPassword{
		User:     "test",
		Password: "test123",
	}, nil, func(addr net.Addr) error {
		conn, err := dx.Dial(addr.Network(), addr.String())
		if err != nil {
			return err
		}
		go func() {
			defer conn.Close()
			io.Copy(conn, conn)
		}()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	conn, err := dr.Dial(dx.LocalAddr.Network(), dx.LocalAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		testConn(t, conn, uuid.NewIdn(4096))
	}
}

func TestSOCKS4BINDRelay(t *testing.T) {
	relay := testRelayServer(t)
	defer relay.Close()

	cfg := &ServerConfig{
		VersionSwitch: DefaultSocksVersionSwitch,
		CMDConfig:     DefaultSocksCMDConfig,
		Socks5AuthCb: S5AuthCb{
			Socks5AuthNOAUTH: DefaultAuthConnCb,
			Socks5AuthPASSWORD: func(conn net.Conn, auth S5AuthPassword) net.Conn {
				return auth.IsEqual2(conn, "test", "test123")
			},
		},
		Socks4AuthCb: S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id S4UserId) (net.Conn, S4IdAuthCode) {
			return id.IsEqual3(conn, S4UserId{1, 2, 3, 4, 5, 6})
		}},
	}
	cfg.CMDConfig.CMDBINDHandler = RelayCMDBINDHandler(func(ctx context.Context) (net.Conn, error) {
		dr := net.Dialer{}
		return dr.DialContext(ctx, relay.Addr().Network(), relay.Addr().String())
	})
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		_ = server.Serve(listen)
	}()
	time.Sleep(1 * time.Second)
	dx := net.Dialer{LocalAddr: &net.TCPAddr{
		IP:   net.IP{127, 0, 0, 1},
		Port: rand.Intn(10000) + 10000,
		Zone: "",
	}}
	dr, err := SOCKS4BIND(listen.Addr().Network(), listen.Addr().String(), S4UserId{1, 2, 3, 4, 5, 6}, nil, func(addr net.Addr) error {
		conn, err := dx.Dial(addr.Network(), addr.String())
		if err != nil {
			return err
		}
		go func() {
			defer conn.Close()
			io.Copy(conn, conn)
		}()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	conn, err := dr.Dial(dx.LocalAddr.Network(), dx.LocalAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		testConn(t, conn, uuid.NewIdn(4096))
	}
}
