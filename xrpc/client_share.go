package xrpc

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"net"
	"sort"
	"sync"
	"time"
)

const ClientNotShareStream = "ClientNotShareStream"

const ClientShareStreamClass = "ClientShareStreamClass"

const ClientShareStreamTmpClass = "ClientShareStreamTmpClass"

func SetClientShareStreamClass(ctx context.Context, class string) context.Context {
	return context.WithValue(ctx, ClientShareStreamClass, class)
}

func SetClientShareStreamTmpClass(ctx context.Context, cfg *ShareStreamConfig) context.Context {
	return context.WithValue(ctx, ClientShareStreamTmpClass, cfg)
}

type RpcFunc func(ctx context.Context, header string, send any, recv any) error

type ShareDialFunc func(ctx context.Context) (net.Conn, error)

func NewShareStreamConfig(class string, mix bool, max, idle int) *ShareStreamConfig {
	return &ShareStreamConfig{
		class: class,
		mix:   mix,
		max:   int64(max),
		idle:  idle,
	}
}

func NewTmpShareStreamConfig(mix bool, max int) *ShareStreamConfig {
	return &ShareStreamConfig{
		class: "_",
		mix:   mix,
		max:   int64(max),
		idle:  0,
	}
}

type ShareStreamConfig struct {
	class string
	mix   bool
	max   int64
	idle  int
}

func (c *Client) newShareManager(fn ShareDialFunc, cfgs ...*ShareStreamConfig) {
	if fn == nil {
		return
	}
	m := make(map[string]*ShareStreamConfig, len(cfgs)+2)
	for _, cfg := range cfgs {
		if _, ok := m[cfg.class]; ok {
			panic("duplicate share stream config: " + cfg.class)
		}
		m[cfg.class] = cfg
	}
	if _, ok := m[""]; !ok {
		m[""] = &ShareStreamConfig{
			class: "",
			mix:   true,
			max:   256,
			idle:  0,
		}
	}
	if _, ok := m[ClientNotShareStream]; !ok {
		m[ClientNotShareStream] = &ShareStreamConfig{
			class: ClientNotShareStream,
			mix:   false,
			max:   0,
			idle:  0,
		}
	}
	c.share = &shareManager{
		c:       c,
		drFunc:  fn,
		cfgMap:  m,
		sChan:   make(chan *ShareStreamConfig, 1),
		sMux:    sync.Mutex{},
		rSelect: false,
		rList:   [2]*ClientSession{},
		rMux:    sync.Mutex{},
		rChan:   make(chan struct{}, 1),
	}
	go c.share.keep()
}

type shareManager struct {
	c       *Client
	drFunc  ShareDialFunc
	cfgMap  map[string]*ShareStreamConfig
	sList   []*streamSession
	sChan   chan *ShareStreamConfig
	sMux    sync.Mutex
	rSelect bool
	rList   [2]*ClientSession
	rMux    sync.Mutex
	rChan   chan struct{}
}

func (sm *shareManager) rpcCallBack(ctx context.Context, fn func(ctx context.Context, fn RpcFunc) error) error {
	sess, err := sm.getShareSessionToRpc(ctx)
	if err != nil {
		return ErrClientShareDialRpcFailed.Errorf(err)
	}
	return fn(sess.Context(), sess.Rpc)
}

func (sm *shareManager) rpc(ctx context.Context, header string, send, recv any) error {
	sess, err := sm.getShareSessionToRpc(ctx)
	if err != nil {
		return ErrClientShareDialRpcFailed.Errorf(err)
	}
	return sess.Rpc(ctx, header, send, recv)
}

func (sm *shareManager) getShareSessionToRpc(ctx context.Context) (*ClientSession, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if sm.c.ctx.Err() != nil {
		return nil, ErrClientClosed
	}
	sm.rMux.Lock()
	defer sm.rMux.Unlock()
	i, j := 0, 1
	if sm.rSelect {
		i, j = 1, 0
	}
	sm.rSelect = !sm.rSelect
	if sm.rList[i] == nil || sm.rList[i].Context().Err() != nil {
		sm.activeR()
	} else {
		return sm.rList[i], nil
	}
	if sm.rList[j] == nil || sm.rList[j].Context().Err() != nil {
		session, err := sm.newSession(ctx)
		if err != nil {
			return nil, err
		}
		sm.rList[j] = session
		return sm.rList[j], nil
	} else {
		return sm.rList[j], nil
	}
}

func (sm *shareManager) stream(ctx context.Context, header string) (Stream, error) {
	session, err := sm.getShareSessionToStream(ctx)
	if err != nil {
		return nil, err
	}
	defer session.streamNum.Add(-1)
	return session.Stream(ctx, header)
}

func (sm *shareManager) recvStream(ctx context.Context, header string, data any) (RecvStream, error) {
	session, err := sm.getShareSessionToStream(ctx)
	if err != nil {
		return nil, err
	}
	defer session.streamNum.Add(-1)
	return session.RecvStream(ctx, header, data)
}

func (sm *shareManager) sendStream(ctx context.Context, header string) (SendStream, error) {
	session, err := sm.getShareSessionToStream(ctx)
	if err != nil {
		return nil, err
	}
	defer session.streamNum.Add(-1)
	return session.SendStream(ctx, header)
}

func (sm *shareManager) reverseRpc(ctx context.Context, header string, data any, route map[string]ClientReverseRpcHandler) error {
	session, err := sm.getShareSessionToStream(ctx)
	if err != nil {
		return err
	}
	defer session.streamNum.Add(-1)
	return session.ReverseRpc(ctx, header, data, route)
}

func (sm *shareManager) getShareSessionToStream(ctx context.Context) (*ClientSession, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if sm.c.ctx.Err() != nil {
		return nil, ErrClientClosed
	}
	var cfg *ShareStreamConfig
	value := ctx.Value(ClientShareStreamClass)
	if value == nil {
		value = ctx.Value(ClientShareStreamTmpClass)
		if value != nil {
			class, ok := value.(*ShareStreamConfig)
			if !ok {
				return nil, ErrInvalidClientShareStreamClass.Errorf("<???>")
			}
			cfg = class
		} else {
			cfg = sm.cfgMap[""]
		}
	} else {
		class, ok := value.(string)
		if !ok {
			return nil, ErrInvalidClientShareStreamClass.Errorf("<???>")
		}
		cfg, ok = sm.cfgMap[class]
		if !ok {
			return nil, ErrInvalidClientShareStreamClass.Errorf(class)
		}
	}
	sm.sMux.Lock()
	defer sm.sMux.Unlock()
	if cfg.max > 0 {
		sort.Slice(sm.sList, func(i, j int) bool {
			return sm.sList[i].streamNum.Load() < sm.sList[j].streamNum.Load()
		})

		for _, session := range sm.sList {
			if session.Context().Err() == nil && session.share &&
				(session.cfg == cfg || (session.cfg.mix && cfg.mix && session.cfg.max == cfg.max)) &&
				session.streamNum.Load() < cfg.max {
				session.streamNum.Add(1)
				session.lastT = time.Now()
				return session.ClientSession, nil
			}
		}
	}
	session, err := sm.newSession(ctx)
	if err != nil {
		return nil, err
	}
	session.streamNum.Add(1)
	sm.sList = append(sm.sList, &streamSession{
		ClientSession: session,
		lastT:         time.Now(),
		cfg:           cfg,
	})
	return session, nil
}

func (sm *shareManager) newSession(ctx context.Context) (*ClientSession, error) {
	tCtx := sm.c.ctx
	if ctx != nil {
		tmpCtx, tmpCl := ctxtool.ContextsWithCancel(sm.c.ctx, ctx)
		defer tmpCl()
		tCtx = tmpCtx
	}
	conn, err := sm.drFunc(tCtx)
	if err != nil {
		return nil, err
	}
	session, err := sm.c.WithConn(tCtx, conn)
	if err != nil {
		return nil, err
	}
	session.share = true
	return session, nil
}

func (sm *shareManager) keep() {
	tk := time.NewTicker(15 * time.Second)
	defer tk.Stop()
	for sm.c.ctx.Err() == nil {
		select {
		case <-tk.C:
			sm.checkS(nil)
		case cfg := <-sm.sChan:
			sm.checkS(cfg)
		case <-sm.rChan:
			sm.checkR()
		case <-sm.c.ctx.Done():
			return
		}
	}
}

func (sm *shareManager) activeR() {
	select {
	case sm.rChan <- struct{}{}:
	default:
	}
}

func (sm *shareManager) activeS(cfg *ShareStreamConfig) {
	select {
	case sm.sChan <- cfg:
	default:
	}
}

func (sm *shareManager) checkR() {
	sm.rMux.Lock()
	sl := []*ClientSession{sm.rList[0], sm.rList[1]}
	sm.rMux.Unlock()
	for i := 0; i < len(sl); i++ {
		if sl[i] == nil || sl[i].Context().Err() != nil {
			session, err := sm.newSession(nil)
			if err != nil {
				return
			}
			sm.rMux.Lock()
			if sm.rList[i] != nil {
				_ = sm.rList[i].Close()
			}
			sm.rList[i] = session
			sm.rMux.Unlock()
		}
	}
}

func (sm *shareManager) checkS(cfg *ShareStreamConfig) {
	var ns *ClientSession
	if cfg != nil {
		var err error
		ns, err = sm.newSession(nil)
		if err != nil {
			return
		}
	}
	sm.sMux.Lock()
	defer sm.sMux.Unlock()
	m := make(map[*ShareStreamConfig]int, len(sm.cfgMap))
	for _, v := range sm.cfgMap {
		m[v] = v.idle
	}
	var sl []*streamSession
	if cfg != nil && ns != nil {
		sm.sList = append(sm.sList, &streamSession{
			ClientSession: ns,
			lastT:         time.Now(),
			cfg:           cfg,
		})
	}
	t := time.Now()
	sl = make([]*streamSession, 0, len(sm.sList))
	for _, session := range sm.sList {
		if session.Context().Err() != nil {
			continue
		}
		if session.streamNum.Load() == 0 {
			if !session.share || session.cfg.max < 1 {
				_ = session.Close()
				continue
			} else if t.Sub(session.lastT) > 60*time.Second {
				if m[session.cfg] <= 0 {
					_ = session.Close()
					continue
				} else {
					m[session.cfg]--
				}
			}
		}
		sl = append(sl, session)
	}
	sm.sList = sl
	for k, i := range m {
		if i > 0 {
			sm.activeS(k)
			break
		}
	}
}

type streamSession struct {
	*ClientSession
	lastT time.Time
	cfg   *ShareStreamConfig
}
