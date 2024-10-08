package httpproxy

import (
	"fmt"
	"github.com/peakedshout/go-pandorasbox/xnet/fasttool"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func Test(t *testing.T) {
	echoHttp, httpAddr, err := fasttool.EchoHttp()
	if err != nil {
		t.Fatal(echoHttp)
	}
	defer echoHttp.Close()

	server, err := NewServer(&ServerConfig{
		ReqAuthCb:   UserInfoAuth("test", "test123"),
		Forward:     nil,
		DialTimeout: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go server.Serve(ln)
	time.Sleep(1 * time.Second)

	parse, err := url.Parse(fmt.Sprintf("http://test:test123@%s", ln.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}

	c := http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(parse)}}
	resp, err := c.Get(httpAddr)
	if err != nil || resp.StatusCode != 200 {
		t.Fatal(err)
	}
}
