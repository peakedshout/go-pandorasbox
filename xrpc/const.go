package xrpc

import "github.com/peakedshout/go-pandorasbox/xmsg"

const (
	SessionAuthInfo = "sessionAuthInfo"
	StreamAuthInfo  = "streamAuthInfo"

	RemotePubNetwork = "remotePubNetwork"
	RemotePubAddress = "remotePubAddress"
	LocalPubNetwork  = "localPubNetwork"
	LocalPubAddress  = "localPubAddress"

	RemotePriNetwork = "remotePriNetwork"
	RemotePriAddress = "remotePriAddress"
	LocalPriNetwork  = "localPriNetwork"
	LocalPriAddress  = "localPriAddress"

	SessionId = "sessionId"

	AuthUserName = "username"
	AuthPassword = "password"
)

const (
	optRpcReq    xmsg.OptType = 11
	optRpcResp   xmsg.OptType = 12
	optRpcFailed xmsg.OptType = 13

	optStreamOpen     xmsg.OptType = 21
	optStreamClose    xmsg.OptType = 22
	optStreamSend     xmsg.OptType = 23
	optStreamRecv     xmsg.OptType = 24
	optStreamFailed   xmsg.OptType = 25
	optStreamPing     xmsg.OptType = 26
	optStreamOpenSend xmsg.OptType = 27
	optStreamOpenRecv xmsg.OptType = 28
	optStreamOpenRRpc xmsg.OptType = 29
)

type typeStream uint8

const (
	typeStreamFullDuplex = typeStream(iota)
	typeStreamSimplexSend
	typeStreamSimplexRecv
)

func (t typeStream) String() string {
	switch t {
	case typeStreamFullDuplex:
		return "fullDuplex"
	case typeStreamSimplexSend:
		return "simplexSend"
	case typeStreamSimplexRecv:
		return "simplexRecv"
	default:
		return "unknown"
	}
}

type Method string

const (
	MethodRpc        Method = "MethodRpc"
	MethodStream     Method = "MethodStream"
	MethodSendStream Method = "MethodSendStream"
	MethodRecvStream Method = "MethodRecvStream"
	MethodReverseRpc Method = "MethodReverseRpc"
)

type Handler interface {
	RpcHandler
	StreamHandler
	SendStreamHandler
	RecvStreamHandler
}
