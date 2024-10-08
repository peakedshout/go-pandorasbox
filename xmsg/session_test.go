package xmsg

import (
	"context"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aesgcm"
	"github.com/peakedshout/go-pandorasbox/protocol/cfcprotocol"
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	key := []byte("00000000000000000000000000000000")
	pc, err := pcrypto.NewPCrypto(aesgcm.PCryptoAes256Gcm, key)
	if err != nil {
		t.Error(err)
	}
	cp := cfcprotocol.NewCFCProtocol(pc)

	ln, err := net.Listen("tcp", "")
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()
	header := "header"
	data := "hello,world!"
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		s1 := NewSession(SessionConfig{
			RWC:      conn,
			Protocol: cp,
			Ctx:      context.Background(),
		})
		defer s1.Close()
		xMsg, _, err := s1.ReadXMsg()
		if err != nil {
			t.Error(err)
		}
		if xMsg.Header() != header || xMsg.Opt() != OptMsg {
			t.Error()
		}
		str := ""
		err = xMsg.Unmarshal(&str)
		if err != nil {
			t.Error(err)
		}
		if str != data {
			t.Error()
		}
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	conn, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	s2 := NewSession(SessionConfig{
		RWC:      conn,
		Protocol: cp,
		Ctx:      context.Background(),
	})
	defer s2.Close()
	_, _, err = s2.SendXMsg(header, 0, OptMsg, data)
	if err != nil {
		t.Error(err)
	}
	wg.Wait()
}

func TestNewSession2(t *testing.T) {
	key := []byte("00000000000000000000000000000000")
	pc, err := pcrypto.NewPCrypto(aesgcm.PCryptoAes256Gcm, key)
	if err != nil {
		t.Error(err)
	}
	cp := cfcprotocol.NewCFCProtocol(pc)

	ln, err := net.Listen("tcp", "")
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()
	header := "header"
	data := "hello,world!"
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		s1 := NewSession(SessionConfig{
			RWC:      conn,
			Protocol: cp,
			Ctx:      context.Background(),
		})
		defer s1.Close()
		for i := 0; i < 10; i++ {
			xMsg, _, err := s1.ReadXMsg()
			if err != nil {
				t.Error(err)
			}
			if xMsg.Header() != header || xMsg.Opt() != OptMsg {
				t.Error()
			}
			str := ""
			err = xMsg.Unmarshal(&str)
			if err != nil {
				t.Error(err)
			}
			if str != data {
				t.Error()
			}
		}
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	conn, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	s2 := NewSession(SessionConfig{
		RWC:      conn,
		Protocol: cp,
		Ctx:      context.Background(),
	})
	defer s2.Close()
	for i := 0; i < 10; i++ {
		_, _, err = s2.SendXMsg(header, 0, OptMsg, data)
		if err != nil {
			t.Error(err)
		}
	}
	wg.Wait()
}

func TestDelay(t *testing.T) {
	key := []byte("00000000000000000000000000000000")
	pc, err := pcrypto.NewPCrypto(aesgcm.PCryptoAes256Gcm, key)
	if err != nil {
		t.Error(err)
	}
	cp := cfcprotocol.NewCFCProtocol(pc)

	ln, err := net.Listen("tcp", "")
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		s1 := NewSession(SessionConfig{
			RWC:      conn,
			Protocol: cp,
			Ctx:      context.Background(),
		})
		defer s1.Close()
		go s1.ReadXMsg()
		wg.Wait()
	}()

	time.Sleep(1 * time.Second)
	conn, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	s2 := NewSession(SessionConfig{
		RWC:      conn,
		Protocol: cp,
		Ctx:      context.Background(),
	})
	defer s2.Close()
	go s2.ReadXMsg()
	time.Sleep(1 * time.Second)
	fmt.Println(s2.Delay(nil))
	wg.Done()
}
