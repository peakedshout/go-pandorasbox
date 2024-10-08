package dnsproxy

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"time"
)

func DefaultFunc(timeout time.Duration, dnsNetwork, dnsAddress string) (func(ctx context.Context, message *dnsmessage.Message) (*dnsmessage.Message, error), error) {
	addr, err := net.ResolveUDPAddr(dnsNetwork, dnsAddress)
	if err != nil {
		return nil, err
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return func(ctx context.Context, message *dnsmessage.Message) (*dnsmessage.Message, error) {
		pack, err := message.Pack()
		if err != nil {
			return nil, err
		}
		config := net.ListenConfig{}
		packet, err := config.ListenPacket(ctx, "udp", "")
		if err != nil {
			return nil, err
		}

		udpConn := packet.(*net.UDPConn)
		xctx, cl := context.WithTimeout(ctx, timeout)
		defer cl()
		ctxtool.GWaitFunc(xctx, func() {
			_ = udpConn.Close()
		})
		_, err = udpConn.WriteToUDP(pack, addr)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, 32*1024)
		var m dnsmessage.Message
		for {
			n, uaddr, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				return nil, err
			}
			if uaddr.String() != addr.String() {
				continue
			}
			err = m.Unpack(buf[:n])
			if err != nil {
				continue
			}
			return &m, nil
		}
	}, nil
}
