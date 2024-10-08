package websocketconn

import "errors"

var (
	ErrBadProtocolVersion   = errors.New("bad protocol version")
	ErrBadScheme            = errors.New("bad scheme")
	ErrBadStatus            = errors.New("bad status")
	ErrBadUpgrade           = errors.New("missing or bad upgrade")
	ErrBadWebSocketOrigin   = errors.New("missing or bad WebSocket-Origin")
	ErrBadWebSocketLocation = errors.New("missing or bad WebSocket-Location")
	ErrBadWebSocketProtocol = errors.New("missing or bad WebSocket-Protocol")
	ErrBadWebSocketVersion  = errors.New("missing or bad WebSocket Version")
	ErrChallengeResponse    = errors.New("mismatch challenge/response")
	ErrBadFrame             = errors.New("bad frame")
	ErrBadFrameBoundary     = errors.New("not on frame boundary")
	ErrNotWebSocket         = errors.New("not websocket protocol")
	ErrBadRequestMethod     = errors.New("bad method")
	ErrNotSupported         = errors.New("not supported")
)
