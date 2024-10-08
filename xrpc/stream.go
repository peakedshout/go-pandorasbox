package xrpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/hjson"
	"github.com/peakedshout/go-pandorasbox/xmsg"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"sync"
	"sync/atomic"
	"time"
)

type streamHandshakeInfo struct {
	AuthInfo AuthInfo
	Data     []byte
}

type StreamHandler func(ctx Stream) error

type serverStream struct {
	sess       *serverSession
	header     string
	id         uint32
	ctx        context.Context
	cl         context.CancelCauseFunc
	read       chan *xmsg.XMsg
	ping       chan struct{}
	closer     sync.Once
	mux        sync.Mutex
	status     bool
	st         typeStream
	monitor    xnetutil.Monitor
	activeTime atomic.Pointer[time.Time]
}

func (ss *serverStream) Id() string {
	return fmt.Sprintf("%s_%d", ss.sess.Id(), ss.id)
}

func (ss *serverStream) Recv(out any) error {
	if ss.st != typeStreamFullDuplex && ss.st != typeStreamSimplexRecv {
		return ErrStreamInvalidAction
	}
	ss.mux.Lock()
	if ss.status {
		ss.mux.Unlock()
		return ErrStreamClosed
	}
	ss.mux.Unlock()
	select {
	case <-ss.ctx.Done():
		return ss.ctx.Err()
	case xMsg, ok := <-ss.read:
		if !ok {
			return ErrStreamClosed
		}
		t := time.Now()
		ss.activeTime.Store(&t)
		if out == nil {
			return nil
		}
		return xMsg.Unmarshal(out)
	}
}

func (ss *serverStream) Send(data any) error {
	if ss.st != typeStreamFullDuplex && ss.st != typeStreamSimplexSend {
		return ErrStreamInvalidAction
	}
	return ss.rawSend(data)
}

func (ss *serverStream) Close() error {
	return ss.close(ErrStreamClosed)
}

func (ss *serverStream) Context() context.Context {
	return ss.ctx
}

func (ss *serverStream) Type() string {
	return ss.st.String()
}

func (ss *serverStream) GetCount() (r, w uint64) {
	return ss.monitor.GetCount()
}

func (ss *serverStream) GetCountView() (r, w string) {
	return ss.monitor.GetCountView()
}

func (ss *serverStream) Speed() (r, w float64) {
	return ss.monitor.Speed()
}

func (ss *serverStream) SpeedView() (r, w string) {
	return ss.monitor.SpeedView()
}

// LifeDuration The duration from the creation of the object to the closure is recorded, not the duration from the time it starts working to the time it stops working.
func (ss *serverStream) LifeDuration() time.Duration {
	return ss.monitor.LifeDuration()
}

func (ss *serverStream) CreateTime() time.Time {
	return ss.monitor.CreateTime()
}

func (ss *serverStream) DeadTime() time.Time {
	return ss.monitor.DeadTime()
}

func (ss *serverStream) GetDelay() time.Duration {
	return ss.monitor.GetDelay()
}

func (ss *serverStream) keepPing(pt time.Duration) {
	defer func() {
		err := ss.ctx.Err()
		var n int
		if err == nil || errors.Is(err, ErrStreamClosed) {
			_, n, _ = ss.sess.RecvXMsg(ss.header, ss.id, optStreamClose, nil)
		} else {
			_, n, _ = ss.sess.RecvXMsg(ss.header, ss.id, optStreamFailed, err)
		}
		ss.monitor.AddCount(0, n)
	}()
	defer ss.close(ErrStreamClosed)
	timeout := time.NewTimer(pt * 4)
	timePing := time.NewTimer(pt)
	defer timeout.Stop()
	defer timePing.Stop()
	for {
		select {
		case <-timePing.C:
			//t := ss.activeTime.Load()
			//if t != nil && t.Sub(time.Now()) > 30*time.Second {
			//	return
			//}
			_, n, err := ss.sess.RecvXMsg(ss.header, ss.id, optStreamPing, nil)
			ss.monitor.AddCount(0, n)
			ss.monitor.RecordSpeed()
			ss.monitor.RecordDelay(ss.GetDelay())
			if err != nil {
				return
			}
			timePing.Reset(pt)
		case <-ss.ping:
			if !timeout.Stop() {
				<-timeout.C
			}
			timeout.Reset(pt * 4)
		case <-timeout.C:
			return
		case <-ss.ctx.Done():
			return
		}
	}
}

func (ss *serverStream) close(err error) error {
	ss.closer.Do(func() {
		ss.mux.Lock()
		ss.status = true
		ss.mux.Unlock()
		ss.cl(ErrStreamClosed)
		err = nil
		ss.monitor.Dead()
		ss.sess.ssMux.Lock()
		delete(ss.sess.streamMap, ss.id)
		if ss.sess.s.cache != nil {
			ss.sess.cacheMap[ss.id] = ss
			ss.sess.s.cache.Duration(ss.sess.s.cacheTime, func() {
				ss.sess.ssMux.Lock()
				delete(ss.sess.cacheMap, ss.id)
				ss.sess.ssMux.Unlock()
			})
		}
		ss.sess.ssMux.Unlock()
	})
	return err
}

func (ss *serverStream) rawSend(data any) error {
	ss.mux.Lock()
	if ss.status {
		ss.mux.Unlock()
		return ErrStreamClosed
	}
	ss.mux.Unlock()
	_, n, err := ss.sess.RecvXMsg(ss.header, ss.id, optStreamRecv, data)
	ss.monitor.AddCount(0, n)
	if err == nil {
		t := time.Now()
		ss.activeTime.Store(&t)
	}
	return err
}

type clientStream struct {
	sess       *ClientSession
	header     string
	id         uint32
	ctx        context.Context
	cl         context.CancelCauseFunc
	read       chan *xmsg.XMsg
	ping       chan struct{}
	closer     sync.Once
	mux        sync.Mutex
	status     bool
	st         typeStream
	opt        xmsg.OptType
	monitor    xnetutil.Monitor
	initialize sync.Once
	initCh     chan error
	activeTime atomic.Pointer[time.Time]
}

func (cs *clientStream) Id() string {
	return fmt.Sprintf("%s_%d", cs.sess.Id(), cs.id)
}

func (cs *clientStream) Recv(out any) error {
	if cs.st != typeStreamFullDuplex && cs.st != typeStreamSimplexRecv {
		return ErrStreamInvalidAction
	}
	cs.mux.Lock()
	if cs.status {
		cs.mux.Unlock()
		return ErrStreamClosed
	}
	cs.mux.Unlock()
	select {
	case <-cs.ctx.Done():
		return cs.ctx.Err()
	case xMsg, ok := <-cs.read:
		if !ok {
			return ErrStreamClosed
		}
		t := time.Now()
		cs.activeTime.Store(&t)
		if out == nil {
			return nil
		}
		return xMsg.Unmarshal(out)
	}
}

// Send If the first message is empty, it is considered an activation signal and is not treated as a message
func (cs *clientStream) Send(data any) error {
	if cs.st != typeStreamFullDuplex && cs.st != typeStreamSimplexSend {
		return ErrStreamInvalidAction
	}
	cs.mux.Lock()
	if cs.status {
		cs.mux.Unlock()
		return ErrStreamClosed
	}
	cs.mux.Unlock()
	isInit, err := cs.toInit(data)
	if isInit {
		if err != nil {
			_ = cs.Close()
			return err
		}
		return nil
	}
	_, n, err := cs.sess.xsess.SendXMsg(cs.header, cs.id, optStreamSend, data)
	cs.monitor.AddCount(0, n)
	if err == nil {
		t := time.Now()
		cs.activeTime.Store(&t)
	}
	return err
}

func (cs *clientStream) Close() error {
	return cs.close(ErrStreamClosed)
}

func (cs *clientStream) Context() context.Context {
	return cs.ctx
}

func (cs *clientStream) Type() string {
	return cs.st.String()
}

func (cs *clientStream) GetCount() (r, w uint64) {
	return cs.monitor.GetCount()
}

func (cs *clientStream) GetCountView() (r, w string) {
	return cs.monitor.GetCountView()
}

func (cs *clientStream) Speed() (r, w float64) {
	return cs.monitor.Speed()
}

func (cs *clientStream) SpeedView() (r, w string) {
	return cs.monitor.SpeedView()
}

// LifeDuration The duration from the creation of the object to the closure is recorded, not the duration from the time it starts working to the time it stops working.
func (cs *clientStream) LifeDuration() time.Duration {
	return cs.monitor.LifeDuration()
}

func (cs *clientStream) CreateTime() time.Time {
	return cs.monitor.CreateTime()
}

func (cs *clientStream) DeadTime() time.Time {
	return cs.monitor.DeadTime()
}

func (cs *clientStream) GetDelay() time.Duration {
	return cs.monitor.GetDelay()
}

func (cs *clientStream) toInit(data any) (bool, error) {
	var err error
	var i bool
	cs.initialize.Do(func() {
		i = true
		send := streamHandshakeInfo{
			AuthInfo: GetStreamAuthInfo(cs.Context()),
		}
		if data != nil {
			send.Data = hjson.MustMarshal(data)
		}
		var n int
		cs.sess.streamMux.Lock()
		cs.id = cs.sess.xsess.GetXMsgId()
		cs.sess.streamMap[cs.id] = cs
		cs.sess.streamMux.Unlock()
		cs.id, n, err = cs.sess.xsess.SendXMsg(cs.header, cs.id, cs.opt, send)
		cs.monitor.AddCount(0, n)
		if err != nil {
			cs.sess.streamMux.Lock()
			delete(cs.sess.streamMap, cs.id)
			cs.sess.streamMux.Unlock()
			_ = cs.Close()
			return
		}
		select {
		case <-cs.ctx.Done():
			err = cs.ctx.Err()
			return
		case err = <-cs.initCh:
			if err == nil {
				go cs.keepPing(cs.sess.c.streamPing)
			}
			return
		}
	})
	return i, err
}

func (cs *clientStream) close(err error) error {
	cs.closer.Do(func() {
		defer cs.sess.streamNum.Add(-1)
		cs.cl(ErrStreamClosed)
		cs.mux.Lock()
		cs.status = true
		cs.mux.Unlock()
		err = nil
		cs.monitor.Dead()
		if cs.id == 0 {
			return
		}
		cs.sess.streamMux.Lock()
		delete(cs.sess.streamMap, cs.id)
		if cs.sess.c.cache != nil {
			cs.sess.cacheMap[cs.id] = cs
			cs.sess.c.cache.Duration(cs.sess.c.cacheTime, func() {
				cs.sess.streamMux.Lock()
				delete(cs.sess.cacheMap, cs.id)
				cs.sess.streamMux.Unlock()
			})
		}
		cs.sess.streamMux.Unlock()
		cs.sess.wg.Done()
	})
	return err
}

func (cs *clientStream) keepPing(pt time.Duration) {
	defer func() {
		err := cs.ctx.Err()
		var n int
		if err == nil || errors.Is(err, ErrStreamClosed) {
			_, n, _ = cs.sess.xsess.SendXMsg(cs.header, cs.id, optStreamClose, nil)
		} else {
			_, n, _ = cs.sess.xsess.SendXMsg(cs.header, cs.id, optStreamFailed, err)
		}
		cs.monitor.AddCount(0, n)
	}()
	defer cs.close(ErrStreamClosed)
	select {
	case <-cs.ctx.Done():
		return
	case <-cs.ping:
	}
	timeout := time.NewTimer(pt * 4)
	timePing := time.NewTimer(pt)
	defer timeout.Stop()
	defer timePing.Stop()
	for {
		select {
		case <-timePing.C:
			//t := cs.activeTime.Load()
			//if t != nil && t.Sub(time.Now()) > 30*time.Second {
			//	return
			//}
			_, n, err := cs.sess.xsess.SendXMsg(cs.header, cs.id, optStreamPing, nil)
			cs.monitor.AddCount(0, n)
			cs.monitor.RecordSpeed()
			cs.monitor.RecordDelay(cs.sess.GetDelay())
			if err != nil {
				return
			}
			timePing.Reset(pt)
		case <-cs.ping:
			if !timeout.Stop() {
				<-timeout.C
			}
			timeout.Reset(pt * 4)
		case <-timeout.C:
			return
		case <-cs.ctx.Done():
			return
		}
	}
}
