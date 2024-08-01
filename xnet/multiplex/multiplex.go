package multiplex

import (
	"context"
	"errors"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/tool/bio"
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"io"
	"time"
)

func NewMultiplex(ctx context.Context, ioTimeout, dialTimeout, listenTimeout time.Duration) *Multiplex {
	if ioTimeout == 0 {
		ioTimeout = 30 * time.Second
	}
	if dialTimeout == 0 {
		dialTimeout = 3 * time.Second
	}
	if listenTimeout == 0 {
		listenTimeout = 3 * time.Second
	}
	m := &Multiplex{
		ioTimeout:     ioTimeout,
		dialTimeout:   dialTimeout,
		listenTimeout: listenTimeout,
		ch:            make(chan *mSession),
	}
	m.ctx, m.cancel = context.WithCancel(ctx)
	return m
}

type SessionFunc func(ctx context.Context) (io.ReadWriteCloser, error)

type Multiplex struct {
	ctx    context.Context
	cancel context.CancelFunc

	ch chan *mSession

	ioTimeout     time.Duration //30s
	dialTimeout   time.Duration //3s
	listenTimeout time.Duration //3s

	unitMap tmap.SyncMap[string, *multiplexUnit]
}

func (m *Multiplex) Stop() {
	m.cancel()
}

func (m *Multiplex) Accept() (io.ReadWriteCloser, error) {
	select {
	case rwc := <-m.ch:
		return rwc, nil
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	}
}

func (m *Multiplex) Listen(sessionFunc SessionFunc) error {
	for {
		rwc, err := sessionFunc(m.ctx)
		if err != nil {
			return err
		}
		go m.Handler(rwc)
	}
}

func (m *Multiplex) Dial(ctx context.Context, sessionFunc SessionFunc, maxConn uint32) (rwc io.ReadWriteCloser, err error) {
	if maxConn == 0 {
		maxConn = 8
	}
	ctxs, cancel := ctxtool.ContextsWithCancel(m.ctx, ctx)
	defer cancel()
	rwc, _ = m.dialScheduler(ctxs, maxConn)
	if rwc != nil {
		if ctxs.Err() != nil {
			_ = rwc.Close()
			return nil, ctxs.Err()
		}
		return rwc, nil
	}
	return m.dialNM(ctxs, sessionFunc)
}

func (m *Multiplex) DialNM(ctx context.Context, sessionFunc SessionFunc) (rwc io.ReadWriteCloser, err error) {
	ctxs, cancel := ctxtool.ContextsWithCancel(m.ctx, ctx)
	defer cancel()
	return m.dialNM(ctxs, sessionFunc)
}

func (m *Multiplex) DialIDle(sessionFunc SessionFunc, idle int, dr time.Duration) error {
	if idle <= 0 {
		return errors.New("invalid idle num")
	}
	if dr == 0 {
		dr = 3 * time.Second
	}
	tr := time.NewTimer(0)
	defer tr.Stop()
	for {
		select {
		case <-tr.C:
			m.dialIDle(sessionFunc, idle)
			tr.Reset(dr)
		case <-m.ctx.Done():
			return m.ctx.Err()
		}
	}
}

func (m *Multiplex) DialDynamicIDle(sessionFunc SessionFunc, idleFunc func() int, dr time.Duration) error {
	if idleFunc == nil {
		return errors.New("invalid idle func")
	}
	if dr == 0 {
		dr = 3 * time.Second
	}
	tr := time.NewTimer(0)
	defer tr.Stop()
	for {
		select {
		case <-tr.C:
			m.dialIDle(sessionFunc, idleFunc())
			tr.Reset(dr)
		case <-m.ctx.Done():
			return m.ctx.Err()
		}
	}
}

func (m *Multiplex) dialIDle(sessionFunc SessionFunc, idle int) {
	for m.ctx.Err() == nil {
		nIdle := 0
		m.unitMap.Range(func(key string, value *multiplexUnit) bool {
			if value.Context().Err() == nil && value.load() == 0 && !value.disable.Load() {
				nIdle++
				if nIdle <= idle {
					value.active()
				}
				if nIdle >= idle*2 {
					value.disable.Store(true)
				}
			}
			return true
		})
		if nIdle < idle {
			rwc, err := sessionFunc(m.ctx)
			if err != nil {
				return
			}
			mio := m.handler(rwc)
			go mio.run()
		} else {
			break
		}
	}
}

func (m *Multiplex) dialNM(ctx context.Context, sessionFunc SessionFunc) (rwc io.ReadWriteCloser, err error) {
	rwc, err = sessionFunc(ctx)
	if err != nil {
		return nil, err
	}
	mio := m.handler(rwc)
	go mio.run()
	return mio.dial(ctx)
}

func (m *Multiplex) Handler(rwc io.ReadWriteCloser) {
	mio := m.handler(rwc)
	mio.run()
}

func (m *Multiplex) handler(rwc io.ReadWriteCloser) *multiplexUnit {
	multiplexIO := bio.NewMultiplexIO(rwc)
	mio := m.newMio(m.ctx, uuid.NewId(1), multiplexIO)
	m.unitMap.Store(mio.id, mio)
	return mio
}

func (m *Multiplex) accept(sess *mSession) bool {
	tr := time.NewTimer(m.listenTimeout)
	defer tr.Stop()
	select {
	case m.ch <- sess:
		return true
	case <-m.ctx.Done():
		return false
	case <-tr.C:
		return false
	}
}

func (m *Multiplex) dialScheduler(ctx context.Context, maxConn uint32) (rwc io.ReadWriteCloser, err error) {
	for ctx.Err() == nil {
		var mio *multiplexUnit
		m.unitMap.Range(func(key string, value *multiplexUnit) bool {
			if value.Context().Err() == nil && !value.disable.Load() && value.load() < maxConn {
				if mio == nil || value.load() < mio.load() {
					mio = value
				}
			}
			return ctx.Err() == nil
		})
		if mio == nil {
			return nil, errors.New("not scheduling")
		}
		if mio.Context().Err() != nil {
			continue
		}
		rwc, err = mio.dial(ctx)
		if err != nil {
			continue
		}
		return rwc, nil
	}
	return nil, ctx.Err()
}

func getOptMsg(b []byte) (opt byte, bs []byte, err error) {
	if len(b) == 0 {
		return 0, nil, errors.New("not opt")
	}
	return b[0], b[1:], nil
}

func setOptMsg(opt byte, bs []byte) []byte {
	return append([]byte{opt}, bs...)
}

const (
	codePing = byte(iota)
	codeMsg
	codeOpen
	codeClose
)
