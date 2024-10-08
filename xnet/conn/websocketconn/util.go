package websocketconn

import (
	"bufio"
	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
	"net/url"
	"strings"
)

const (
	ProtocolVersionHybi13    = 13
	ProtocolVersionHybi      = ProtocolVersionHybi13
	SupportedProtocolVersion = "13"

	ContinuationFrame = 0
	TextFrame         = 1
	BinaryFrame       = 2
	CloseFrame        = 8
	PingFrame         = 9
	PongFrame         = 10
	UnknownFrame      = 255

	DefaultMaxPayloadBytes = 32 << 20 // 32MB
)

func (wc *WebsocketConn) newConn() {
	wc.frameReaderFactory = hybiFrameReaderFactory{wc.reader}
	wc.frameWriterFactory = hybiFrameWriterFactory{
		Writer:         wc.writer,
		needMaskingKey: wc.isClient,
	}
	wc.payloadType = BinaryFrame
	wc.defaultCloseStatus = closeStatusNormal
	wc.frameHandler = &hybiFrameHandler{conn: wc}
}

func (wc *WebsocketConn) handShakeHandle() error {
	wc.reader = bufio.NewReader(wc.conn)
	wc.writer = bufio.NewWriter(wc.conn)
	if wc.isClient {
		err := wc.handShakeClient()
		if err != nil {
			return err
		}
	} else {
		err := wc.handShakeServer()
		if err != nil {
			return err
		}
	}
	return nil
}

func (wc *WebsocketConn) handShakeServer() error {
	request, err := http.ReadRequest(wc.reader)
	if err != nil {
		return err
	}
	code, err := wc.readHandshake(request)
	if err != nil {
		return err
	}
	if err == ErrBadWebSocketVersion {
		fmt.Fprintf(wc.writer, "HTTP/1.1 %03d %s\r\n", code, http.StatusText(code))
		fmt.Fprintf(wc.writer, "Sec-WebSocket-Version: %s\r\n", SupportedProtocolVersion)
		wc.writer.WriteString("\r\n")
		wc.writer.WriteString(err.Error())
		wc.writer.Flush()
		return err
	}
	if err != nil {
		fmt.Fprintf(wc.writer, "HTTP/1.1 %03d %s\r\n", code, http.StatusText(code))
		wc.writer.WriteString("\r\n")
		wc.writer.WriteString(err.Error())
		wc.writer.Flush()
		return err
	}
	err = wc.acceptHandshake()
	if err != nil {
		code = http.StatusBadRequest
		fmt.Fprintf(wc.writer, "HTTP/1.1 %03d %s\r\n", code, http.StatusText(code))
		wc.writer.WriteString("\r\n")
		wc.writer.Flush()
		return err
	}
	wc.newConn()
	return nil
}

func (wc *WebsocketConn) readHandshake(req *http.Request) (code int, err error) {
	wc.version = websocket.ProtocolVersionHybi13
	if req.Method != "GET" {
		return http.StatusMethodNotAllowed, ErrBadRequestMethod
	}

	if strings.ToLower(req.Header.Get("Upgrade")) != "websocket" ||
		!strings.Contains(strings.ToLower(req.Header.Get("Connection")), "upgrade") {
		return http.StatusBadRequest, ErrNotWebSocket
	}

	key := req.Header.Get("Sec-Websocket-Key")
	if key == "" {
		return http.StatusBadRequest, ErrChallengeResponse
	}
	version := req.Header.Get("Sec-Websocket-Version")
	switch version {
	case "13":
		wc.version = ProtocolVersionHybi13
	default:
		return http.StatusBadRequest, ErrBadWebSocketVersion
	}
	var scheme string
	if req.TLS != nil {
		scheme = "wss"
	} else {
		scheme = "ws"
	}
	wc.location, err = url.ParseRequestURI(scheme + "://" + req.Host + req.URL.RequestURI())
	if err != nil {
		return http.StatusBadRequest, err
	}
	protocol := strings.TrimSpace(req.Header.Get("Sec-Websocket-Protocol"))
	if protocol != "" {
		protocols := strings.Split(protocol, ",")
		for i := 0; i < len(protocols); i++ {
			wc.protocol = append(wc.protocol, strings.TrimSpace(protocols[i]))
		}
	}
	wc.accept, err = getNonceAccept([]byte(key))
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusSwitchingProtocols, nil
}

func (wc *WebsocketConn) acceptHandshake() (err error) {
	if len(wc.protocol) > 0 {
		if len(wc.protocol) != 1 {
			return ErrBadWebSocketProtocol
		}
	}
	wc.writer.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	wc.writer.WriteString("Upgrade: websocket\r\n")
	wc.writer.WriteString("Connection: Upgrade\r\n")
	wc.writer.WriteString("Sec-WebSocket-Accept: " + string(wc.accept) + "\r\n")
	if len(wc.protocol) > 0 {
		wc.writer.WriteString("Sec-WebSocket-Protocol: " + wc.protocol[0] + "\r\n")
	}
	if wc.header != nil {
		err := wc.header.WriteSubset(wc.writer, handshakeHeader)
		if err != nil {
			return err
		}
	}
	wc.writer.WriteString("\r\n")
	return wc.writer.Flush()
}

func (wc *WebsocketConn) handShakeClient() (err error) {
	wc.writer.WriteString("GET " + wc.location.RequestURI() + " HTTP/1.1\r\n")

	wc.writer.WriteString("Host: " + removeZone(wc.location.Host) + "\r\n")
	wc.writer.WriteString("Upgrade: websocket\r\n")
	wc.writer.WriteString("Connection: Upgrade\r\n")
	nonce := generateNonce()
	if wc.handshakeData != nil {
		nonce = []byte(wc.handshakeData["key"])
	}
	wc.writer.WriteString("Sec-WebSocket-Key: " + string(nonce) + "\r\n")
	wc.writer.WriteString("Origin: " + strings.ToLower(wc.origin.String()) + "\r\n")

	if wc.version != ProtocolVersionHybi13 {
		return ErrBadProtocolVersion
	}

	wc.writer.WriteString("Sec-WebSocket-Version: " + fmt.Sprintf("%d", wc.version) + "\r\n")
	if len(wc.protocol) > 0 {
		wc.writer.WriteString("Sec-WebSocket-Protocol: " + strings.Join(wc.protocol, ", ") + "\r\n")
	}
	err = wc.header.WriteSubset(wc.writer, handshakeHeader)
	if err != nil {
		return err
	}

	wc.writer.WriteString("\r\n")
	if err = wc.writer.Flush(); err != nil {
		return err
	}

	resp, err := http.ReadResponse(wc.reader, &http.Request{Method: "GET"})
	if err != nil {
		return err
	}
	if resp.StatusCode != 101 {
		return ErrBadStatus
	}
	if strings.ToLower(resp.Header.Get("Upgrade")) != "websocket" ||
		strings.ToLower(resp.Header.Get("Connection")) != "upgrade" {
		return ErrBadUpgrade
	}
	expectedAccept, err := getNonceAccept(nonce)
	if err != nil {
		return err
	}
	if resp.Header.Get("Sec-WebSocket-Accept") != string(expectedAccept) {
		return ErrChallengeResponse
	}
	if resp.Header.Get("Sec-WebSocket-Extensions") != "" {
		return ErrUnsupportedExtensions
	}
	offeredProtocol := resp.Header.Get("Sec-WebSocket-Protocol")
	if offeredProtocol != "" {
		protocolMatched := false
		for i := 0; i < len(wc.protocol); i++ {
			if wc.protocol[i] == offeredProtocol {
				protocolMatched = true
				break
			}
		}
		if !protocolMatched {
			return ErrBadWebSocketProtocol
		}
		wc.protocol = []string{offeredProtocol}
	}
	wc.newConn()
	return nil
}

func (wc *WebsocketConn) defaultClientCfg() {
	scheme := "http://"
	wscheme := "ws://"
	if wc.tlsCfg != nil {
		scheme = "https://"
		wscheme = "wss://"
	}
	wc.version = ProtocolVersionHybi13
	wc.location, _ = url.ParseRequestURI(fmt.Sprintf("%s%s/", scheme, wc.raddr.String()))
	wc.origin, _ = url.ParseRequestURI(fmt.Sprintf("%s%s/", wscheme, wc.raddr.String()))
}
