package xhttp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/hjson"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"io"
	"net"
	"net/http"
	"testing"
)

func TestNewServer(t *testing.T) {
	server := NewServer(&Config{})
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	fmt.Println("ln:", listen.Addr().String())
	server.Set("test", func(context *Context) error {
		var str string
		err2 := context.Bind(&str)
		if err2 != nil {
			return err2
		}
		_, _ = context.WriteString(str)
		return nil
	})
	defer server.Close()
	go server.Serve(listen)
	data := uuid.NewId(1)
	resp, err := http.Post("http://"+listen.Addr().String()+"/test", "", bytes.NewBufferString(hjson.MustMarshalStr(data)))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp.Request.URL)
	if resp.StatusCode != http.StatusOK {
		t.Fatal()
	}
	all, _ := io.ReadAll(resp.Body)
	if string(all) != data {
		t.Fatal()
	}
}

func TestNewClient(t *testing.T) {
	server := NewServer(&Config{})
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	fmt.Println("ln:", listen.Addr().String())
	server.Set("test", func(context *Context) error {
		var str string
		err2 := context.Bind(&str)
		if err2 != nil {
			return err2
		}
		_, _ = context.WriteString(str)
		return nil
	})
	defer server.Close()
	go server.Serve(listen)
	data := uuid.NewId(1)
	client := NewClient(&ClientConfig{
		Host: listen.Addr().String(),
	})
	bs, err := client.CallBytes(context.Background(), "test", data)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != data {
		t.Fatal()
	}
}
