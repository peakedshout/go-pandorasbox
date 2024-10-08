package xrpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aesgcm"
	"github.com/peakedshout/go-pandorasbox/protocol/cfcprotocol"
	"github.com/peakedshout/go-pandorasbox/tool/expired"
	"github.com/peakedshout/go-pandorasbox/tool/hjson"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/peakedshout/go-pandorasbox/xmsg"
	"github.com/peakedshout/go-pandorasbox/xnet/xflow"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
	"sort"
	"sync"
	"time"
)

func UPAuthCallback(u string, p string) func(info AuthInfo) (AuthInfo, error) {
	return func(info AuthInfo) (AuthInfo, error) {
		cu := info.Get(AuthUserName)
		cp := info.Get(AuthPassword)
		if u != cu || p != cp {
			return nil, ErrAuthVerificationFailed
		}
		return make(AuthInfo), nil
	}
}

type ServerConfig struct {
	Ctx                      context.Context
	KeepLive                 time.Duration
	HandshakeTimeout         time.Duration
	StreamPing               time.Duration
	SwitchNetworkSpeedTicker bool
	SessionAuthCallback      func(info AuthInfo) (AuthInfo, error)
	StreamAuthCallback       func(info AuthInfo) (AuthInfo, error)
	CryptoList               []*CryptoConfig
	Upgrader                 xnetutil.Upgrader
	CacheTime                time.Duration
}

func NewServer(sc *ServerConfig) *Server {
	s := new(Server)
	ctx := sc.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.sessionAuthCb = sc.SessionAuthCallback
	s.streamAuthCb = sc.StreamAuthCallback
	if sc.KeepLive >= 1*time.Second {
		s.keepLive = sc.KeepLive
	} else {
		s.keepLive = 5 * time.Second
	}
	if sc.HandshakeTimeout >= 1*time.Second {
		s.handshakeTimeout = sc.HandshakeTimeout
	} else {
		s.handshakeTimeout = 5 * time.Second
	}
	if sc.StreamPing > 5*time.Second {
		s.streamPing = sc.StreamPing
	} else {
		s.streamPing = 5 * time.Second
	}
	if len(sc.CryptoList) != 0 {
		s.crypto = sc.CryptoList
		sort.Slice(s.crypto, func(i, j int) bool {
			return s.crypto[i].Priority < s.crypto[j].Priority
		})
	} else {
		s.crypto = []*CryptoConfig{{Crypto: pcrypto.CryptoPlaintext}}
	}
	if sc.CacheTime > 1*time.Second {
		s.cacheTime = sc.CacheTime
		s.cache = expired.NewTODO(expired.Init(s.ctx, 1))
	}
	s.rpcRoute = make(map[string]RpcHandler)
	s.ssRoute = make(map[string]StreamHandler)
	s.rsRoute = make(map[string]SendStreamHandler)
	s.srRoute = make(map[string]RecvStreamHandler)
	s.rrRoute = make(map[string]ReverseRpcHandler)
	return s
}

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc

	keepLive                 time.Duration
	handshakeTimeout         time.Duration
	streamPing               time.Duration
	switchNetworkSpeedTicker bool

	sessionAuthCb func(info AuthInfo) (AuthInfo, error)
	streamAuthCb  func(info AuthInfo) (AuthInfo, error)
	crypto        []*CryptoConfig
	upgrader      xnetutil.Upgrader

	rpcRoute map[string]RpcHandler
	ssRoute  map[string]StreamHandler
	rsRoute  map[string]SendStreamHandler
	srRoute  map[string]RecvStreamHandler
	rrRoute  map[string]ReverseRpcHandler

	running bool
	closer  sync.Once
	mux     sync.Mutex
	disable bool
	wg      sync.WaitGroup

	sessMap   tmap.SyncMap[string, *serverSession]
	cacheTime time.Duration
	cache     *expired.TODO
}

func (s *Server) MustAddHandler(header string, handler any) {
	switch h := handler.(type) {
	case RpcHandler:
		s.MustAddRpcHandler(header, h)
	case StreamHandler:
		s.MustAddStreamHandler(header, h)
	case SendStreamHandler:
		s.MustAddSendStreamHandler(header, h)
	case RecvStreamHandler:
		s.MustAddRecvStreamHandler(header, h)
	case ReverseRpcHandler:
		s.MustAddReverseRpcHandler(header, h)
	case func(ctx Rpc) (any, error):
		s.MustAddRpcHandler(header, h)
	case func(ctx Stream) error:
		s.MustAddStreamHandler(header, h)
	case func(ctx SendStream) error:
		s.MustAddSendStreamHandler(header, h)
	case func(ctx RecvStream) (any, error):
		s.MustAddRecvStreamHandler(header, h)
	case func(ReverseRpc) error:
		s.MustAddReverseRpcHandler(header, h)
	default:
		panic("invalid method " + fmt.Sprintf("%T", handler))
	}
}

func (s *Server) MustAddRpcHandler(header string, handler RpcHandler) *Server {
	err := s.AddRpcHandler(header, handler)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Server) MustAddStreamHandler(header string, handler StreamHandler) *Server {
	err := s.AddStreamHandler(header, handler)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Server) MustAddSendStreamHandler(header string, handler SendStreamHandler) *Server {
	err := s.AddSendStreamHandler(header, handler)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Server) MustAddRecvStreamHandler(header string, handler RecvStreamHandler) *Server {
	err := s.AddRecvStreamHandler(header, handler)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Server) MustAddReverseRpcHandler(header string, handler ReverseRpcHandler) *Server {
	err := s.AddReverseRpcHandler(header, handler)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Server) AddRpcHandler(header string, handler RpcHandler) error {
	s.mux.Lock()
	if s.disable {
		s.mux.Unlock()
		return ErrServerClosed
	}
	if s.running {
		s.mux.Unlock()
		return ErrServerRunning
	}
	s.mux.Unlock()
	s.rpcRoute[header] = handler
	return nil
}

func (s *Server) AddStreamHandler(header string, handler StreamHandler) error {
	s.mux.Lock()
	if s.disable {
		s.mux.Unlock()
		return ErrServerClosed
	}
	if s.running {
		s.mux.Unlock()
		return ErrServerRunning
	}
	s.mux.Unlock()
	s.ssRoute[header] = handler
	return nil
}

func (s *Server) AddSendStreamHandler(header string, handler SendStreamHandler) error {
	s.mux.Lock()
	if s.disable {
		s.mux.Unlock()
		return ErrServerClosed
	}
	if s.running {
		s.mux.Unlock()
		return ErrServerRunning
	}
	s.mux.Unlock()
	s.rsRoute[header] = handler
	return nil
}

func (s *Server) AddRecvStreamHandler(header string, handler RecvStreamHandler) error {
	s.mux.Lock()
	if s.disable {
		s.mux.Unlock()
		return ErrServerClosed
	}
	if s.running {
		s.mux.Unlock()
		return ErrServerRunning
	}
	s.mux.Unlock()
	s.srRoute[header] = handler
	return nil
}

func (s *Server) AddReverseRpcHandler(header string, handler ReverseRpcHandler) error {
	s.mux.Lock()
	if s.disable {
		s.mux.Unlock()
		return ErrServerClosed
	}
	if s.running {
		s.mux.Unlock()
		return ErrServerRunning
	}
	s.mux.Unlock()
	s.rrRoute[header] = handler
	return nil
}

func (s *Server) Serve(ln net.Listener) error {
	s.mux.Lock()
	if s.disable {
		s.mux.Unlock()
		return ErrServerClosed
	}
	s.running = true
	s.wg.Add(1)
	defer s.wg.Done()
	s.mux.Unlock()
	ctx, cl := context.WithCancel(s.ctx)
	defer cl()
	ctxtool.GWaitFunc(ctx, func() {
		_ = ln.Close()
	})
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.serveConn(conn)
	}
}

func (s *Server) Close() error {
	err := ErrServerClosed
	s.closer.Do(func() {
		s.cancel()
		s.mux.Lock()
		s.disable = true
		s.mux.Unlock()
		s.wg.Wait()
		err = nil
	})
	return err
}

func (s *Server) Context() context.Context {
	return s.ctx
}

func (s *Server) serveConn(conn net.Conn) {
	_ = conn.SetDeadline(time.Now().Add(s.handshakeTimeout))
	defer conn.Close()
	if s.upgrader != nil {
		var err error
		conn, err = s.upgrader.UpgradeContext(s.ctx, conn)
		if err != nil {
			return
		}
	}
	if s.switchNetworkSpeedTicker {
		conn, _ = xflow.FlowUpgrader().Upgrade(conn)
	}
	pCrypto, err := s.handleSelectCrypto(conn)
	if err != nil {
		return
	}
	sc := xmsg.SessionConfig{
		RWC:      conn,
		Protocol: cfcprotocol.NewCFCProtocol(pCrypto),
		KeepLive: s.keepLive,
		Ctx:      s.ctx,
		Flag:     xmsg.FlagOne,
	}
	var authInfo AuthInfo
	err = sc.Protocol.Decode(conn, &authInfo)
	if err != nil {
		return
	}
	authInfo.connSet(true, conn)
	if s.sessionAuthCb != nil {
		info, err := s.sessionAuthCb(authInfo)
		if err != nil {
			return
		}
		if info == nil {
			info = make(AuthInfo)
		}
		info.connSet(true, conn)
		err = sc.Protocol.Encode(conn, info)
		if err != nil {
			return
		}
	} else {
		err = sc.Protocol.Encode(conn, make(AuthInfo).connSet(true, conn))
		if err != nil {
			return
		}
	}
	sc.Ctx = SetSessionAuthInfo(s.ctx, authInfo.connSet(true, conn))
	session := xmsg.NewSession(sc)
	_ = conn.SetDeadline(time.Time{})
	authInfo.Set(SessionId, hjson.MustMarshalStr(session.Id()))
	ss := &serverSession{
		s:          s,
		RawSession: session,
		cacheMap:   make(map[uint32]*serverStream),
		streamMap:  make(map[uint32]*serverStream),
	}
	s.sessMap.Store(ss.Id(), ss)
	if s.cache == nil {
		defer s.sessMap.Delete(ss.Id())
	} else {
		defer s.cache.Duration(s.cacheTime, func() {
			s.sessMap.Delete(ss.Id())
		})
	}
	s.handleXMsg(ss)
}

func (s *Server) handleSelectCrypto(conn net.Conn) (pcrypto.PCrypto, error) {
	var cryptoNameList []string
	err := cfcprotocol.CFCPlaintext.Decode(conn, &cryptoNameList)
	if err != nil {
		return nil, err
	}
	l := len(cryptoNameList)
	if l == 0 {
		return nil, ErrServerSelectCrypto.Errorf("nil crypto list")
	}
	m := make(map[string]bool, l)
	for _, one := range cryptoNameList {
		m[one] = true
	}
	var pc *CryptoConfig
	for _, one := range s.crypto {
		if m[one.String()] {
			pc = one
			break
		}
	}
	if pc == nil {
		return nil, ErrServerSelectCrypto.Errorf("There is no supported crypto")
	}
	err = cfcprotocol.CFCPlaintext.Encode(conn, pc.String())
	if err != nil {
		return nil, err
	}
	if !pc.Crypto.IsSymmetric() {
		var cb []byte
		err = cfcprotocol.NewCFCProtocol(pc.Crypto).Decode(conn, &cb)
		if err != nil {
			return nil, err
		}
		b := mhash.ToHash([]byte(uuid.NewIdn(64)))
		err = cfcprotocol.CFCPlaintext.Encode(conn, b)
		if err != nil {
			return nil, err
		}
		key := mhash.ToHash(bytes.Join([][]byte{cb, b}, nil))
		c, err := pcrypto.NewPCrypto(aesgcm.PCryptoAes256Gcm, key)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	return pc.Crypto, nil
}

func (s *Server) handleXMsg(session *serverSession) {
	defer session.Close()
	for {
		xMsg, n, err := session.ReadXMsg()
		if err != nil {
			return
		}
		switch xMsg.Opt() {
		case optRpcReq:
			s.handleRpc(session, xMsg)
		case optStreamOpen:
			s.handleOpenStream(session, xMsg, n)
		case optStreamOpenRecv:
			s.handleOpenStreamRecv(session, xMsg, n)
		case optStreamOpenSend:
			s.handleOpenStreamSend(session, xMsg, n)
		case optStreamOpenRRpc:
			s.handleRRpc(session, xMsg, n)
		case optStreamClose, optStreamFailed:
			s.handleCloseStream(session, xMsg, n)
		case optStreamSend, optStreamPing:
			s.handleSendStream(session, xMsg, n)
		default:
			continue
		}
	}
}

func (s *Server) handleRpc(session *serverSession, xMsg *xmsg.XMsg) {
	handler, ok := s.rpcRoute[xMsg.Header()]
	if !ok {
		_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optRpcFailed, ErrInvalidCall.Errorf(xMsg.Header()))
	}
	go func() {
		tmpCtx, cl := context.WithCancel(session.Context())
		ctx := &rpcContext{
			ctx:  tmpCtx,
			xMsg: xMsg,
		}
		data, err := handler(ctx)
		cl()
		if err != nil {
			_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optRpcFailed, err)
			return
		}
		_, _, err = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optRpcResp, data)
		if err != nil {
			_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optRpcFailed, err)
			return
		}
	}()
}

func (s *Server) handleOpenStream(session *serverSession, xMsg *xmsg.XMsg, r int) {
	handler, ok := s.ssRoute[xMsg.Header()]
	if !ok {
		_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optStreamFailed, ErrInvalidCall.Errorf(xMsg.Header()))
		return
	}
	ctx, err := session.newStream(xMsg, optStreamOpen, r)
	if err != nil {
		_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optStreamFailed, err)
		return
	}
	//s.wg.Add(1)
	go func() {
		//defer s.wg.Done()
		err := handler(ctx)
		_ = ctx.close(err)
	}()
}

func (s *Server) handleOpenStreamSend(session *serverSession, xMsg *xmsg.XMsg, r int) {
	handler, ok := s.rsRoute[xMsg.Header()]
	if !ok {
		_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optStreamFailed, ErrInvalidCall.Errorf(xMsg.Header()))
		return
	}
	ctx, err := session.newStream(xMsg, optStreamOpenSend, r)
	if err != nil {
		_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optStreamFailed, err)
		return
	}
	//s.wg.Add(1)
	go func() {
		//defer s.wg.Done()
		err := handler(&sendStreamContext{
			xMsg:   <-ctx.read,
			stream: ctx,
		})
		_ = ctx.close(err)
	}()
}

func (s *Server) handleOpenStreamRecv(session *serverSession, xMsg *xmsg.XMsg, r int) {
	handler, ok := s.srRoute[xMsg.Header()]
	if !ok {
		_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optStreamFailed, ErrInvalidCall.Errorf(xMsg.Header()))
		return
	}
	ctx, err := session.newStream(xMsg, optStreamOpenRecv, r)
	if err != nil {
		_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optStreamFailed, err)
		return
	}
	//s.wg.Add(1)
	go func() {
		//defer s.wg.Done()
		data, err := handler(&recvServerStream{ctx})
		if err != nil {
			_ = ctx.close(err)
		} else {
			_ = ctx.close(ctx.rawSend(data))
		}
	}()
}

func (s *Server) handleRRpc(session *serverSession, xMsg *xmsg.XMsg, r int) {
	handler, ok := s.rrRoute[xMsg.Header()]
	if !ok {
		_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optStreamFailed, ErrInvalidCall.Errorf(xMsg.Header()))
		return
	}
	ctx, err := session.newStream(xMsg, optStreamOpenRRpc, r)
	if err != nil {
		_, _, _ = session.RecvXMsg(xMsg.Header(), xMsg.Id(), optStreamFailed, err)
		return
	}
	//s.wg.Add(1)
	go func() {
		//defer s.wg.Done()
		rctx := &serverReverseRpcContext{
			xMsg:   <-ctx.read,
			stream: ctx,
			rpcMap: make(map[uint32]chan *rRpcMsg),
		}
		go rctx.handleXMsg()
		err := handler(rctx)
		_ = ctx.close(err)
	}()
}

func (s *Server) handleSendStream(session *serverSession, xMsg *xmsg.XMsg, r int) {
	session.ssMux.Lock()
	sc, ok := session.streamMap[xMsg.Id()]
	if !ok || sc.header != xMsg.Header() {
		sc = nil
	}
	session.ssMux.Unlock()
	if sc == nil || sc.ctx.Err() != nil {
		return
	}
	sc.monitor.AddCount(r, 0)
	switch xMsg.Opt() {
	case optStreamSend:
		if sc.st != typeStreamFullDuplex && sc.st != typeStreamSimplexRecv {
			return
		}
		select {
		case <-session.Context().Done():
			_ = sc.close(ErrStreamClosed)
		case <-sc.ctx.Done():
		case sc.read <- xMsg:
		}
	case optStreamPing:
		select {
		case <-session.Context().Done():
			sc.close(ErrStreamClosed)
		case <-sc.ctx.Done():
		case sc.ping <- struct{}{}:
		}
	}
}

func (s *Server) handleCloseStream(session *serverSession, xMsg *xmsg.XMsg, r int) {
	session.ssMux.Lock()
	sc, ok := session.streamMap[xMsg.Id()]
	if !ok || sc.header != xMsg.Header() {
		sc = nil
	}
	if sc == nil || sc.ctx.Err() != nil {
		session.ssMux.Unlock()
		return
	}
	session.ssMux.Unlock()
	sc.monitor.AddCount(r, 0)
	switch xMsg.Opt() {
	case optStreamFailed:
		var str string
		err := xMsg.Unmarshal(&str)
		if err != nil {
			_ = sc.close(err)
			return
		}
		_ = sc.close(errors.New(str))
	case optStreamClose:
		_ = sc.close(ErrStreamClosed)
	}
}

func (s *Server) GetDelay(sid string) time.Duration {
	sess, ok := s.sessMap.Load(sid)
	if ok {
		return sess.GetDelay()
	}
	return 0
}

type serverSession struct {
	s *Server

	*xmsg.RawSession
	ssMux sync.Mutex
	//ssMap    map[uint32]*streamContext
	cacheMap map[uint32]*serverStream

	streamMap map[uint32]*serverStream
}

func (ss *serverSession) GetDelay() time.Duration {
	return ss.RawSession.GetDelay()
}

func (ss *serverSession) newStream(xMsg *xmsg.XMsg, opt xmsg.OptType, r int) (*serverStream, error) {
	st := typeStreamFullDuplex
	switch opt {
	case optStreamOpenSend:
		st = typeStreamSimplexSend
	case optStreamOpenRecv:
		st = typeStreamSimplexRecv
	}
	monitor := xnetutil.NewMonitor()
	stream := &serverStream{
		sess:    ss,
		header:  xMsg.Header(),
		id:      xMsg.Id(),
		ctx:     nil,
		cl:      nil,
		read:    make(chan *xmsg.XMsg, 1),
		ping:    make(chan struct{}),
		st:      st,
		monitor: monitor,
	}
	// handshake
	var info streamHandshakeInfo
	err := xMsg.Unmarshal(&info)
	if err != nil {
		return nil, err
	}
	if info.AuthInfo == nil {
		info.AuthInfo = make(AuthInfo)
	}
	var sendInfo AuthInfo
	if ss.s.streamAuthCb != nil {
		sendInfo, err = ss.s.streamAuthCb(info.AuthInfo)
		if err != nil {
			return nil, err
		}
		if sendInfo == nil {
			sendInfo = make(AuthInfo)
		}
	}

	// data
	if len(info.Data) == 0 {
		_ = xMsg.Marshal(nil)
	} else {
		_ = xMsg.Marshal(xmsg.NoBytes(info.Data))
	}
	if xMsg.NilData() && (opt == optStreamOpen || opt == optStreamRecv) {
	} else {
		stream.read <- xMsg
	}

	streamCtx := SetStreamAuthInfo(ss.Context(), info.AuthInfo)
	stream.ctx, stream.cl = context.WithCancelCause(streamCtx)
	ss.ssMux.Lock()
	old, ok := ss.streamMap[xMsg.Id()]
	if ok {
		_ = old.close(ErrStreamClosed)
	}
	ss.streamMap[xMsg.Id()] = stream
	ss.ssMux.Unlock()
	_, n, err := ss.RecvXMsg(xMsg.Header(), xMsg.Id(), opt, sendInfo)
	monitor.AddCount(r, n)
	if err != nil {
		_ = stream.Close()
		return nil, err
	}
	go stream.keepPing(ss.s.streamPing)
	return stream, nil
}
