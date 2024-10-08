package xrpc

import (
	"bytes"
	"context"
	"errors"
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
	"sync/atomic"
	"time"
)

func BuildUPAuth(u string, p string) AuthInfo {
	m := make(AuthInfo)
	m.Set(AuthUserName, u)
	m.Set(AuthPassword, p)
	return m
}

type ClientConfig struct {
	Ctx                      context.Context
	KeepLive                 time.Duration
	HandshakeTimeout         time.Duration
	StreamPing               time.Duration
	SwitchNetworkSpeedTicker bool
	SessionAuthInfo          AuthInfo
	StreamAuthInfo           AuthInfo
	CryptoList               []*CryptoConfig
	Upgrader                 xnetutil.Upgrader
	CacheTime                time.Duration
	ShareDialFunc            ShareDialFunc
	ShareStreamConfigList    []*ShareStreamConfig
}

func NewClient(cc *ClientConfig) *Client {
	c := new(Client)
	ctx := cc.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.sessionAuthInfo = cc.SessionAuthInfo
	c.streamAuthInfo = cc.StreamAuthInfo
	if cc.KeepLive >= 1*time.Second {
		c.keepLive = cc.KeepLive
	} else {
		c.keepLive = 5 * time.Second
	}
	if cc.HandshakeTimeout >= 1*time.Second {
		c.handshakeTimeout = cc.HandshakeTimeout
	} else {
		c.handshakeTimeout = 5 * time.Second
	}
	if cc.StreamPing > 5*time.Second {
		c.streamPing = cc.StreamPing
	} else {
		c.streamPing = 5 * time.Second
	}
	if len(cc.CryptoList) != 0 {
		c.crypto = cc.CryptoList
		sort.Slice(c.crypto, func(i, j int) bool {
			return c.crypto[i].Priority < c.crypto[j].Priority
		})
	} else {
		c.crypto = []*CryptoConfig{{Crypto: pcrypto.CryptoPlaintext}}
	}
	if cc.CacheTime > 1*time.Second {
		c.cacheTime = cc.CacheTime
		c.cache = expired.NewTODO(expired.Init(c.ctx, 1))
	}
	c.newShareManager(cc.ShareDialFunc, cc.ShareStreamConfigList...)
	return c
}

type Client struct {
	ctx    context.Context
	cancel context.CancelFunc

	keepLive                 time.Duration
	handshakeTimeout         time.Duration
	streamPing               time.Duration
	switchNetworkSpeedTicker bool

	sessionAuthInfo AuthInfo
	streamAuthInfo  AuthInfo
	crypto          []*CryptoConfig
	upgrader        xnetutil.Upgrader

	closer  sync.Once
	mux     sync.Mutex
	disable bool
	wg      sync.WaitGroup

	sessMap   tmap.SyncMap[string, *ClientSession]
	cacheTime time.Duration
	cache     *expired.TODO

	share *shareManager
}

func (c *Client) Context() context.Context {
	return c.ctx
}

func (c *Client) DialContext(ctx context.Context, dr xnetutil.Dialer, network string, addr string) (*ClientSession, error) {
	c.mux.Lock()
	if c.disable {
		c.mux.Unlock()
		return nil, ErrClientClosed
	}
	c.wg.Add(1)
	c.mux.Unlock()
	ctxs, cl := ctxtool.ContextsWithCancel(c.ctx, ctx)
	defer cl()
	conn, err := dr.DialContext(ctxs, network, addr)
	if err != nil {
		c.wg.Done()
		return nil, err
	}
	defer func() {
		if err != nil {
			c.wg.Done()
			_ = conn.Close()
		}
	}()
	cs, err := c.handleConn(ctxs, conn)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

func (c *Client) WithConn(ctx context.Context, conn net.Conn) (*ClientSession, error) {
	c.mux.Lock()
	if c.disable {
		c.mux.Unlock()
		return nil, ErrClientClosed
	}
	c.wg.Add(1)
	c.mux.Unlock()
	tmpCtx, tmpCl := context.WithCancel(ctx)
	defer tmpCl()
	cs, err := c.handleConn(tmpCtx, conn)
	if err != nil {
		c.wg.Done()
		_ = conn.Close()
		return nil, err
	}
	return cs, nil
}

func (c *Client) Close() error {
	err := ErrClientClosed
	c.closer.Do(func() {
		c.cancel()
		c.mux.Lock()
		c.disable = true
		c.mux.Unlock()
		c.wg.Wait()
		err = nil
	})
	return err
}

func (c *Client) Rpc(ctx context.Context, header string, send, recv any) error {
	if c.share == nil {
		return ErrClientNilShareDialMethod
	}
	return c.share.rpc(ctx, header, send, recv)
}

func (c *Client) Stream(ctx context.Context, header string) (Stream, error) {
	if c.share == nil {
		return nil, ErrClientNilShareDialMethod
	}
	return c.share.stream(ctx, header)
}

func (c *Client) RecvStream(ctx context.Context, header string, data any) (RecvStream, error) {
	if c.share == nil {
		return nil, ErrClientNilShareDialMethod
	}
	return c.share.recvStream(ctx, header, data)
}

func (c *Client) SendStream(ctx context.Context, header string) (SendStream, error) {
	if c.share == nil {
		return nil, ErrClientNilShareDialMethod
	}
	return c.share.sendStream(ctx, header)
}

func (c *Client) ReverseRpc(ctx context.Context, header string, data any, route map[string]ClientReverseRpcHandler) error {
	if c.share == nil {
		return ErrClientNilShareDialMethod
	}
	return c.share.reverseRpc(ctx, header, data, route)
}

func (c *Client) RpcCallback(ctx context.Context, fn func(ctx context.Context, fn RpcFunc) error) error {
	if c.share == nil {
		return ErrClientNilShareDialMethod
	}
	return c.share.rpcCallBack(ctx, fn)
}

func (c *Client) handleConn(ctx context.Context, conn net.Conn) (*ClientSession, error) {
	var err error
	k := true
	ctxtool.GWaitFunc(ctx, func() {
		if k {
			_ = conn.Close()
		}
	})
	_ = conn.SetDeadline(time.Now().Add(c.handshakeTimeout))
	if c.upgrader != nil {
		conn, err = c.upgrader.UpgradeContext(ctx, conn)
		if err != nil {
			return nil, err
		}
	}
	if c.switchNetworkSpeedTicker {
		conn, _ = xflow.FlowUpgrader().Upgrade(conn)
	}
	pCrypto, err := c.handleSelectCrypto(conn)
	if err != nil {
		return nil, err
	}
	sc := xmsg.SessionConfig{
		RWC:      conn,
		Protocol: cfcprotocol.NewCFCProtocol(pCrypto),
		KeepLive: c.keepLive,
		Ctx:      c.ctx,
		Flag:     xmsg.FlagOne,
	}
	auth := GetSessionAuthInfo(ctx)
	authInfo := make(AuthInfo, len(c.sessionAuthInfo)+len(auth)+4)
	for key, value := range c.sessionAuthInfo {
		authInfo[key] = value
	}
	for key, value := range auth {
		authInfo[key] = value
	}
	authInfo.connSet(false, conn)
	err = sc.Protocol.Encode(conn, authInfo)
	if err != nil {
		return nil, err
	}
	err = sc.Protocol.Decode(conn, &authInfo)
	if err != nil {
		return nil, err
	}
	authInfo.connSet(false, conn)
	sc.Ctx = SetSessionAuthInfo(c.ctx, authInfo)
	session := xmsg.NewSession(sc)
	_ = conn.SetDeadline(time.Time{})
	authInfo.Set(SessionId, hjson.MustMarshalStr(session.Id()))
	cs := &ClientSession{
		c:         c,
		xsess:     session,
		rpcMap:    make(map[uint32]chan *xmsg.XMsg),
		streamMap: make(map[uint32]*clientStream),
		cacheMap:  make(map[uint32]*clientStream),
	}
	go cs.handleXMsg()
	k = false
	c.sessMap.Store(cs.Id(), cs)
	return cs, nil
}

func (c *Client) handleSelectCrypto(conn net.Conn) (pcrypto.PCrypto, error) {
	sl := make([]string, 0, len(c.crypto))
	for _, one := range c.crypto {
		sl = append(sl, one.String())
	}
	err := cfcprotocol.CFCPlaintext.Encode(conn, sl)
	if err != nil {
		return nil, err
	}
	var cryptoName string
	err = cfcprotocol.CFCPlaintext.Decode(conn, &cryptoName)
	if err != nil {
		return nil, err
	}
	var pc *CryptoConfig
	for _, config := range c.crypto {
		if config.String() == cryptoName {
			pc = config
			return config.Crypto, nil
		}
	}
	if pc == nil {
		return nil, ErrClientSelectCrypto.Errorf("There is no supported crypto")
	}
	if pc.Crypto.IsSymmetric() {
		b := mhash.ToHash([]byte(uuid.NewIdn(64)))
		err = cfcprotocol.NewCFCProtocol(pc.Crypto).Encode(conn, b)
		if err != nil {
			return nil, err
		}
		var cb []byte
		err = cfcprotocol.CFCPlaintext.Decode(conn, &cb)
		if err != nil {
			return nil, err
		}
		key := mhash.ToHash(bytes.Join([][]byte{b, cb}, nil))
		p, err := pcrypto.NewPCrypto(aesgcm.PCryptoAes256Gcm, key)
		if err != nil {
			return nil, err
		}
		return p, nil
	}
	return pc.Crypto, nil
}

type ClientSession struct {
	c     *Client
	xsess *xmsg.RawSession
	share bool

	rpcMux    sync.Mutex
	rpcMap    map[uint32]chan *xmsg.XMsg
	streamNum atomic.Int64
	streamMux sync.Mutex
	streamMap map[uint32]*clientStream
	cacheMap  map[uint32]*clientStream

	closer  sync.Once
	mux     sync.Mutex
	disable bool
	wg      sync.WaitGroup
}

func (cs *ClientSession) Context() context.Context {
	return cs.xsess.Context()
}

func (cs *ClientSession) Close() error {
	err := ErrClientSessionClosed
	cs.closer.Do(func() {
		cs.mux.Lock()
		cs.disable = true
		cs.mux.Unlock()
		_ = cs.xsess.Close()
		cs.wg.Wait()
		if cs.c.cache == nil {
			cs.c.sessMap.Delete(cs.Id())
		} else {
			cs.c.cache.Duration(cs.c.cacheTime, func() {
				cs.c.sessMap.Delete(cs.Id())
			})
		}
		cs.c.wg.Done()
		err = nil
	})
	return err
}

func (cs *ClientSession) Rpc(ctx context.Context, header string, send, recv any) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	cs.mux.Lock()
	if cs.disable {
		cs.mux.Unlock()
		return ErrClientSessionClosed
	}
	cs.wg.Add(1)
	defer cs.wg.Done()
	cs.mux.Unlock()
	id, _, err := cs.xsess.SendXMsg(header, 0, optRpcReq, send)
	if err != nil {
		return err
	}
	ch := make(chan *xmsg.XMsg, 1)
	cs.setRpc(id, ch)
	defer cs.delRpc(id)
	select {
	case xMsg := <-ch:
		if xMsg.Opt() == optRpcFailed {
			var errStr string
			err = xMsg.Unmarshal(&errStr)
			if err != nil {
				return err
			}
			return errors.New(errStr)
		}
		if recv == nil {
			return nil
		}
		return xMsg.Unmarshal(recv)
	case <-ctx.Done():
		return ctx.Err()
	case <-cs.xsess.Context().Done():
		return cs.xsess.Context().Err()
	}
}

func (cs *ClientSession) Stream(ctx context.Context, header string) (Stream, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	cs.mux.Lock()
	if cs.disable {
		cs.mux.Unlock()
		return nil, ErrClientSessionClosed
	}
	cs.wg.Add(1)
	cs.streamNum.Add(1)
	cs.mux.Unlock()
	s := cs.newStream(ctx, header, optStreamOpen)
	return s, nil
}

func (cs *ClientSession) RecvStream(ctx context.Context, header string, data any) (RecvStream, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	cs.mux.Lock()
	if cs.disable {
		cs.mux.Unlock()
		return nil, ErrClientSessionClosed
	}
	cs.wg.Add(1)
	cs.streamNum.Add(1)
	cs.mux.Unlock()
	rs := cs.newStream(ctx, header, optStreamOpenSend)
	_, err := rs.toInit(data)
	if err != nil {
		_ = rs.Close()
		return nil, err
	}
	return &recvClientStream{cs: rs}, nil
}

func (cs *ClientSession) SendStream(ctx context.Context, header string) (SendStream, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	cs.mux.Lock()
	if cs.disable {
		cs.mux.Unlock()
		return nil, ErrClientSessionClosed
	}
	cs.wg.Add(1)
	cs.streamNum.Add(1)
	cs.mux.Unlock()
	s := cs.newStream(ctx, header, optStreamOpenRecv)
	ss := &sendClientStream{
		cs: s,
	}
	go ss.run()
	return ss, nil
}

func (cs *ClientSession) ReverseRpc(ctx context.Context, header string, data any, route map[string]ClientReverseRpcHandler) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	cs.mux.Lock()
	if cs.disable {
		cs.mux.Unlock()
		return ErrClientSessionClosed
	}
	cs.wg.Add(1)
	cs.streamNum.Add(1)
	cs.mux.Unlock()
	s := cs.newStream(ctx, header, optStreamOpenRRpc)
	defer s.Close()
	err := s.Send(data)
	if err != nil {
		return err
	}
	for {
		msg := new(rRpcMsg)
		err := s.Recv(msg)
		if err != nil {
			return err
		}
		handler, ok := route[msg.Header]
		if ok {
			go func() {
				tmpCtx, cl := context.WithCancel(s.ctx)
				bs, err := handler(&clientReverseRpcContext{
					ctx: tmpCtx,
					msg: msg,
				})
				cl()
				if err != nil {
					_ = s.Send(&rRpcMsg{
						Header: msg.Header,
						Id:     msg.Id,
						Type:   rMsgTypeErr,
						Data:   []byte(err.Error()),
					})
				} else {
					_ = s.Send(&rRpcMsg{
						Header: msg.Header,
						Id:     msg.Id,
						Type:   rMsgTypeMsg,
						Data:   hjson.MustMarshal(bs),
					})
				}
			}()
		}
	}
}

func (cs *ClientSession) Id() string {
	return cs.xsess.Id()
}

func (cs *ClientSession) GetDelay() time.Duration {
	return cs.xsess.GetDelay()
}

func (cs *ClientSession) GetCount() (r, w uint64) {
	return cs.xsess.GetCount()
}

func (cs *ClientSession) Speed() (r, w float64) {
	return cs.xsess.Speed()
}

func (cs *ClientSession) SpeedView() (r, w string) {
	return cs.xsess.SpeedView()
}

func (cs *ClientSession) LifeDuration() time.Duration {
	return cs.xsess.LifeDuration()
}

func (cs *ClientSession) CreateTime() time.Time {
	return cs.xsess.CreateTime()
}

func (cs *ClientSession) DeadTime() time.Time {
	return cs.xsess.DeadTime()
}

func (cs *ClientSession) newStream(ctx context.Context, header string, opt xmsg.OptType) *clientStream {
	st := typeStreamFullDuplex
	switch opt {
	case optStreamOpenSend:
		st = typeStreamSimplexRecv
	case optStreamOpenRecv:
		st = typeStreamSimplexSend
	}
	s := &clientStream{
		sess:    cs,
		header:  header,
		id:      0,
		read:    make(chan *xmsg.XMsg),
		ping:    make(chan struct{}),
		status:  false,
		st:      st,
		opt:     opt,
		monitor: xnetutil.NewMonitor(),
		initCh:  make(chan error, 1),
	}

	auth := GetStreamAuthInfo(ctx)
	authInfo := make(AuthInfo, len(cs.c.streamAuthInfo)+len(auth))
	for key, value := range cs.c.streamAuthInfo {
		authInfo[key] = value
	}
	for key, value := range auth {
		authInfo[key] = value
	}
	streamCtx := SetStreamAuthInfo(cs.xsess.Context(), authInfo)
	s.ctx, s.cl = ctxtool.ContextsWithCancelCause(streamCtx, ctx)
	return s
}

func (cs *ClientSession) handleXMsg() {
	defer cs.Close()
	for {
		xMsg, n, err := cs.xsess.ReadXMsg()
		if err != nil {
			return
		}
		switch xMsg.Opt() {
		case optRpcResp, optRpcFailed:
			cs.handleRpc(xMsg)
		case optStreamOpen, optStreamPing,
			optStreamOpenRecv, optStreamOpenSend, optStreamOpenRRpc,
			optStreamRecv, optStreamClose, optStreamFailed:
			cs.handleStream(xMsg, n)
		default:
			continue
		}
	}
}

func (cs *ClientSession) handleRpc(xMsg *xmsg.XMsg) {
	cs.rpcMux.Lock()
	defer cs.rpcMux.Unlock()
	ch, ok := cs.rpcMap[xMsg.Id()]
	if ok {
		ch <- xMsg
		delete(cs.rpcMap, xMsg.Id())
	}
}

func (cs *ClientSession) setRpc(sid uint32, ch chan *xmsg.XMsg) {
	cs.rpcMux.Lock()
	defer cs.rpcMux.Unlock()
	cs.rpcMap[sid] = ch
}

func (cs *ClientSession) delRpc(sid uint32) {
	cs.rpcMux.Lock()
	defer cs.rpcMux.Unlock()
	delete(cs.rpcMap, sid)
}

func (cs *ClientSession) handleStream(xMsg *xmsg.XMsg, r int) {
	cs.streamMux.Lock()
	s, ok := cs.streamMap[xMsg.Id()]
	cs.streamMux.Unlock()
	if ok {
		s.monitor.AddCount(r, 0)
		switch xMsg.Opt() {
		case optStreamOpen, optStreamOpenSend, optStreamOpenRecv, optStreamOpenRRpc:
			var info AuthInfo
			err := xMsg.Unmarshal(&info)
			if err == nil {
				authInfo := GetStreamAuthInfo(s.Context())
				for k, v := range info {
					authInfo[k] = v
				}
			}
			select {
			case s.initCh <- err:
			default:
			}
		case optStreamPing:
			select {
			case <-cs.xsess.Context().Done():
				_ = s.Close()
			case <-s.ctx.Done():
			case s.ping <- struct{}{}:
			}
		case optStreamRecv:
			select {
			case <-cs.xsess.Context().Done():
				_ = s.Close()
			case <-s.ctx.Done():
			case s.read <- xMsg:
			}
		case optStreamClose:
			_ = s.Close()
		case optStreamFailed:
			var str string
			err := xMsg.Unmarshal(&str)
			if err == nil {
				err = errors.New(str)
			}
			_ = s.close(err)
			select {
			case s.initCh <- err:
			default:
			}
		}
	}
}
