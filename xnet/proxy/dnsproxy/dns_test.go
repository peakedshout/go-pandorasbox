package dnsproxy

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestDNS(t *testing.T) {
	rcb, err := DefaultFunc(10*time.Second, "udp", "8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(rcb, 0)
	defer server.Close()
	ln, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go server.Serve(ln)
	time.Sleep(1 * time.Second)

	crcb, err := DefaultFunc(10*time.Second, ln.LocalAddr().Network(), ln.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(crcb)
	if err != nil {
		t.Fatal(err)
	}
	ips, err := client.LookupIP("www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ips)
}
