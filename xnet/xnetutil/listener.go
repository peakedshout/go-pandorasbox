package xnetutil

import (
	"context"
	"net"
)

type ListenerConfig interface {
	Listen(network string, address string) (net.Listener, error)
	ListenContext(ctx context.Context, network string, address string) (net.Listener, error)
}

type PacketListenerConfig interface {
	ListenPacket(network string, address string) (net.PacketConn, error)
	ListenPacketContext(ctx context.Context, network string, address string) (net.PacketConn, error)
}

type defaultListenerConfig struct {
	*net.ListenConfig
}

func NewDefaultListenerConfig(lc *net.ListenConfig) ListenerConfig {
	if lc == nil {
		lc = &net.ListenConfig{}
	}
	return &defaultListenerConfig{ListenConfig: lc}
}

func (dlc *defaultListenerConfig) Listen(network string, address string) (net.Listener, error) {
	return dlc.ListenContext(context.Background(), network, address)
}

func (dlc *defaultListenerConfig) ListenContext(ctx context.Context, network string, address string) (net.Listener, error) {
	return dlc.ListenConfig.Listen(ctx, network, address)
}

func (dlc *defaultListenerConfig) ListenPacket(network string, address string) (net.PacketConn, error) {
	return dlc.ListenConfig.ListenPacket(context.Background(), network, address)
}

func (dlc *defaultListenerConfig) ListenPacketContext(ctx context.Context, network string, address string) (net.PacketConn, error) {
	return dlc.ListenConfig.ListenPacket(ctx, network, address)
}
