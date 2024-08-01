package fakehttpconn

import (
	"bytes"
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/pcrypto/rsa"
	"github.com/peakedshout/go-pandorasbox/tool/bio"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestFakeHttpConn1(t *testing.T) {
	data := []byte("12313123")
	var wg sync.WaitGroup
	wg.Add(1)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		hconn := newFakeHttpConn(false, conn, nil)
		defer hconn.Close()
		buf := make([]byte, len(data))
		n, err := io.ReadFull(hconn, buf)
		if !bytes.Equal(buf[:n], data) {
			t.Error()
		}
		_, err = hconn.Write(data)
		if err != nil {
			t.Error(err)
		}
		wg.Wait()
	}()
	time.Sleep(1 * time.Second)
	conn, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	hconn := newFakeHttpConn(true, conn, nil)
	defer hconn.Close()
	_, err = hconn.Write(data)
	if err != nil {
		t.Error(err)
	}
	buf := make([]byte, len(data))
	n, err := io.ReadFull(hconn, buf)
	if !bytes.Equal(buf[:n], data) {
		t.Error()
	}
	wg.Done()
}

func TestFakeHttpConn2(t *testing.T) {
	data := []byte("12313123")
	var wg sync.WaitGroup
	wg.Add(1)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		hconn := newFakeHttpConn(false, conn, nil)
		defer hconn.Close()
		buf := make([]byte, len(data))
		n, err := io.ReadFull(hconn, buf)
		if !bytes.Equal(buf[:n], data) {
			t.Error()
		}
		_, err = hconn.Write(data)
		if err != nil {
			t.Error(err)
		}
		wg.Wait()
	}()
	time.Sleep(1 * time.Second)

	var c http.Client
	req, err := http.NewRequest(http.MethodGet, "http://"+ln.Addr().String(), bio.NewNoCloseBody(data))
	if err != nil {
		t.Error(err)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Error(err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	if !bytes.Equal(b, data) {
		t.Error()
	}
	wg.Done()
}

func TestFakeHttpConn3(t *testing.T) {
	data := []byte("12313123")

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()
	go func() {
		h := &handle{
			data: data,
			fn: func(r bool, err error) {
				if !r || err != nil {
					ln.Close()
					t.Error(err)
				}
			},
		}
		http.Serve(ln, h)
	}()
	time.Sleep(1 * time.Second)
	conn, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	hconn := newFakeHttpConn(true, conn, nil)
	defer hconn.Close()
	_, err = hconn.Write(data)
	if err != nil {
		t.Error(err)
	}
	buf := make([]byte, len(data))
	n, err := io.ReadFull(hconn, buf)
	if !bytes.Equal(buf[:n], data) {
		t.Error()
	}
}

func TestFakeHttpConn4(t *testing.T) {
	cert, key, err := rsa.PCryptoRsaCert.GenRsaCert(1024, nil)
	if err != nil {
		t.Error(err)
	}
	c, err := tls.X509KeyPair(cert, key)
	if err != nil {
		t.Error(err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{c}, InsecureSkipVerify: true}

	data := []byte("12313123")
	var wg sync.WaitGroup
	wg.Add(1)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		hconn := newFakeHttpConn(false, conn, cfg)
		defer hconn.Close()
		buf := make([]byte, len(data))
		n, err := io.ReadFull(hconn, buf)
		if !bytes.Equal(buf[:n], data) {
			t.Error()
		}
		_, err = hconn.Write(data)
		if err != nil {
			t.Error(err)
		}
		wg.Wait()
	}()
	time.Sleep(1 * time.Second)
	conn, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	hconn := newFakeHttpConn(true, conn, cfg)
	defer hconn.Close()
	_, err = hconn.Write(data)
	if err != nil {
		t.Error(err)
	}
	buf := make([]byte, len(data))
	n, err := io.ReadFull(hconn, buf)
	if !bytes.Equal(buf[:n], data) {
		t.Error()
	}
	wg.Done()
}

func TestFakeHttpConn5(t *testing.T) {
	cert, key, err := rsa.PCryptoRsaCert.GenRsaCert(1024, nil)
	if err != nil {
		t.Error(err)
	}
	c, err := tls.X509KeyPair(cert, key)
	if err != nil {
		t.Error(err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{c}, InsecureSkipVerify: true}

	data := []byte("12313123")
	var wg sync.WaitGroup
	wg.Add(1)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		hconn := newFakeHttpConn(false, conn, cfg)
		defer hconn.Close()
		buf := make([]byte, len(data))
		n, err := io.ReadFull(hconn, buf)
		if !bytes.Equal(buf[:n], data) {
			t.Error()
		}
		_, err = hconn.Write(data)
		if err != nil {
			t.Error(err)
		}
		wg.Wait()
	}()
	time.Sleep(1 * time.Second)

	var client http.Client
	client.Transport = &http.Transport{TLSClientConfig: cfg}

	req, err := http.NewRequest(http.MethodGet, "https://"+ln.Addr().String(), bio.NewNoCloseBody(data))
	if err != nil {
		t.Error(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	if !bytes.Equal(b, data) {
		t.Error()
	}
	wg.Done()
}

func TestFakeHttpConn6(t *testing.T) {
	cert, key, err := rsa.PCryptoRsaCert.GenRsaCert(1024, nil)
	if err != nil {
		t.Error(err)
	}
	c, err := tls.X509KeyPair(cert, key)
	if err != nil {
		t.Error(err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{c}, InsecureSkipVerify: true}

	data := []byte("12313123")

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	ln = tls.NewListener(ln, cfg)
	defer ln.Close()
	go func() {
		h := &handle{
			data: data,
			fn: func(r bool, err error) {
				if !r || err != nil {
					ln.Close()
					t.Error(err)
				}
			},
		}
		http.Serve(ln, h)
	}()
	time.Sleep(1 * time.Second)
	conn, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	hconn := newFakeHttpConn(true, conn, cfg)
	defer hconn.Close()
	_, err = hconn.Write(data)
	if err != nil {
		t.Error(err)
	}
	buf := make([]byte, len(data))
	n, err := io.ReadFull(hconn, buf)
	if !bytes.Equal(buf[:n], data) {
		t.Error()
	}
}

type handle struct {
	data []byte
	fn   func(r bool, err error)
}

func (h *handle) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	b, err := io.ReadAll(request.Body)
	if err != nil {
		h.fn(false, err)
		return
	}
	r := bytes.Equal(b, h.data)
	if !r {
		h.fn(r, nil)
		return
	}
	_, err = writer.Write(h.data)
	if err != nil {
		h.fn(false, err)
		return
	}
}
