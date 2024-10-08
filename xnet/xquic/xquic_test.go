package xquic

import (
	"bytes"
	"context"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/peakedshout/go-pandorasbox/xnet/conn/packetconn"
	"github.com/peakedshout/go-pandorasbox/xnet/conn/quicconn"
	"github.com/peakedshout/go-pandorasbox/xnet/xudp"
	"github.com/quic-go/quic-go"
	"net"
	"testing"
)

func TestLD(t *testing.T) {
	listen, err := Listen("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	go func() {
		conn, err := listen.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
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
	}()
	conn, err := Dial(listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = conn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}

func TestListen(t *testing.T) {
	listen, err := Listen("udp", "127.0.0.1:0")
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
	}()
	connection, err := quic.DialAddr(context.Background(), listen.Addr().String(), _defaultQuicTlsConfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := quicconn.NewConn(true, connection)
	defer conn.Close()
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = conn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}

func TestDial(t *testing.T) {
	listener, err := quic.ListenAddr("127.0.0.1:0", _defaultQuicTlsConfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		connection, err := listener.Accept(context.Background())
		if err != nil {
			t.Error(err)
			return
		}
		conn := quicconn.NewConn(false, connection)
		defer conn.Close()
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
	}()
	conn, err := Dial(listener.Addr().Network(), listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = conn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}

func TestTD2(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer udpConn.Close()
	addr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	udpConn2, err := net.ListenUDP("udp", addr2)
	if err != nil {
		t.Fatal(err)
	}
	defer udpConn2.Close()
	xconn, err := xudp.UDPUpgrader().Upgrade(udpConn)
	if err != nil {
		t.Fatal(err)
	}
	xconn2 := packetconn.NewPacketConn(udpConn2, addr)
	if err != nil {
		t.Fatal(err)
	}
	listen, err := NewQuicListenConfig(xconn.(net.PacketConn), _defaultQuicTlsConfg).Listen("udp", "127.0.0.1:0")
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
	}()
	conn, err := NewDialer(xconn2.(net.PacketConn), _defaultQuicTlsConfg).Dial(listen.Addr().Network(), listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = conn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}
