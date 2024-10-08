package multiplex

import (
	"context"
	"errors"
	"github.com/peakedshout/go-pandorasbox/tool/bio"
	"github.com/peakedshout/go-pandorasbox/tool/task"
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

func (m *Multiplex) newMio(ctx context.Context, id string, mio *bio.MultiplexIO) *multiplexUnit {
	mi := &multiplexUnit{id: id, m: m, MultiplexIO: mio, activeCh: make(chan struct{}, 1)}
	mi.ctx, mi.cancel = context.WithCancel(ctx)
	mi.tc = task.NewTaskCtx[mSession](mi.ctx)
	go mi.check()
	return mi
}

type multiplexUnit struct {
	id string

	disable    atomic.Bool
	activeTime atomic.Pointer[time.Time]
	activeCh   chan struct{}

	m *Multiplex
	*bio.MultiplexIO
	mux    sync.Mutex
	count  uint32
	ctx    context.Context
	cancel context.CancelFunc

	once sync.Once

	sessionMap tmap.SyncMap[uint32, *mSession]
	ring       uint32

	tc *task.TaskCtx[mSession]
}

func (m *multiplexUnit) dial(ctx context.Context) (io.ReadWriteCloser, error) {
	if m.disable.Load() {
		return nil, errors.New("unit is disable")
	}
	id := m.getRing()
	tmpCtx, tmpCl := context.WithTimeout(ctx, m.m.dialTimeout)
	defer tmpCl()
	return m.tc.RegisterTask(tmpCtx, id, func() error {
		return m.WriteMsg(setOptMsg(codeOpen, nil), 0, id)
	})
}

func (m *multiplexUnit) Context() context.Context {
	return m.ctx
}

func (m *multiplexUnit) add(i uint32) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.count += i
}

func (m *multiplexUnit) sub(i uint32) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.count -= i
}

func (m *multiplexUnit) load() uint32 {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.count
}

func (m *multiplexUnit) run() {
	defer m.close()
	for {
		msg, rid, lid, err := m.ReadMsg()
		if err != nil {
			return
		}
		opt, bs, err := getOptMsg(msg)
		if err != nil {
			continue
		}
		m.active()
		switch opt {
		case codeOpen:
			if rid == 0 {
				id := m.getRing()
				session := newSession(m.ctx, func() {
					m.sub(1)
					m.sessionMap.Delete(id)
					_ = m.WriteMsg(setOptMsg(codeClose, nil), lid, id)
				}, func(bytes []byte) error {
					m.active()
					return m.WriteMsgWithPreSize(bytes, []byte{codeMsg}, lid, id, buffSize)
				})
				session.ping()
				m.add(1)
				m.sessionMap.Store(id, session)
				err = m.WriteMsg(setOptMsg(codeOpen, nil), lid, id)
				if err != nil {
					_ = session.Close()
				}
				if !m.m.accept(session) {
					_ = session.Close()
				}
			} else {
				session := newSession(m.ctx, func() {
					m.sub(1)
					m.sessionMap.Delete(lid)
					_ = m.WriteMsg(setOptMsg(codeClose, nil), lid, rid)
				}, func(bytes []byte) error {
					m.active()
					return m.WriteMsgWithPreSize(bytes, []byte{codeMsg}, lid, rid, buffSize)
				})
				session.ping()
				m.add(1)
				m.sessionMap.Store(rid, session)
				if !m.tc.CallBack(rid, session, nil) {
					_ = session.Close()
				}
			}
		case codePing:
			value, ok := m.sessionMap.Load(rid)
			if ok {
				value.ping()
			}
		case codeMsg:
			value, ok := m.sessionMap.Load(rid)
			if ok {
				err = value.writeCh(bs)
				if err != nil {
					_ = value.Close()
				}
			}
		default:
			value, ok := m.sessionMap.Load(rid)
			if ok {
				_ = value.Close()
			}
		}
	}
}

func (m *multiplexUnit) close() {
	m.once.Do(func() {
		m.cancel()
		_ = m.MultiplexIO.Close()
		m.sessionMap.Range(func(key uint32, value *mSession) bool {
			_ = value.Close()
			return true
		})
		m.m.unitMap.Delete(m.id)
	})
}

func (m *multiplexUnit) check() {
	m.active()
	defer m.close()
	td := m.m.ioTimeout // default 30s
	tr := time.NewTimer(td)
	defer tr.Stop()
	for {
		select {
		case <-tr.C:
			t := time.Now().Add(-m.m.ioTimeout)
			if m.activeTime.Load().Before(t) {
				m.disable.Store(true)
			}
			if m.disable.Load() && m.load() == 0 {
				return
			}
			m.sessionMap.Range(func(key uint32, value *mSession) bool {
				if value.activeTime.Load().Before(t) {
					_ = value.Close()
				}
				return true
			})
			tr.Reset(td)
		case <-m.ctx.Done():
			return
		case <-m.activeCh:
			t := time.Now()
			m.activeTime.Store(&t)
		}
	}
}

func (m *multiplexUnit) getRing() uint32 {
	id := atomic.AddUint32(&m.ring, 1)
	if id == 0 {
		id = atomic.AddUint32(&m.ring, 1)
	}
	return id
}

func (m *multiplexUnit) active() {
	select {
	case m.activeCh <- struct{}{}:
	default:
	}
}
