package xrpc

import "github.com/peakedshout/go-pandorasbox/tool/xerror"

var (
	ErrServerSelectCrypto  = xerror.New("server select crypto bad: %s")
	ErrClientSelectCrypto  = xerror.New("client select crypto bad: %s")
	ErrServerClosed        = xerror.New("server closed")
	ErrServerRunning       = xerror.New("server running")
	ErrClientSessionClosed = xerror.New("client session closed")
	ErrClientClosed        = xerror.New("client closed")

	ErrClientNilShareDialMethod      = xerror.New("client nil share dial method")
	ErrClientShareDialRpcFailed      = xerror.New("client share dial rpc failed: %w")
	ErrInvalidClientShareStreamClass = xerror.New("invalid client share stream class: %v")

	ErrRRpcClosed = xerror.New("reverse rpc closed")

	ErrStreamClosed        = xerror.New("stream closed")
	ErrStreamInvalidAction = xerror.New("stream invalid action")
	ErrInvalidCall         = xerror.New("invalid call: %v")

	ErrAuthVerificationFailed = xerror.New("auth verification failed")
)
