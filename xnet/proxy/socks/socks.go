package socks

import "net"

func Serve(ln net.Listener, cfg *ServerConfig) error {
	server, err := NewServer(cfg)
	if err != nil {
		return err
	}
	defer server.Close()
	return server.Serve(ln)
}

func ListenAndServe(network string, addr string, cfg *ServerConfig) error {
	server, err := NewServer(cfg)
	if err != nil {
		return err
	}
	defer server.Close()
	return server.ListenAndServe(network, addr)
}
