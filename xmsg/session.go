package xmsg

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/protocol"
	"github.com/peakedshout/go-pandorasbox/protocol/jsonprotocol"
	"github.com/peakedshout/go-pandorasbox/tool/ticker"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"io"
	"sync"
	"time"
)

type SessionConfig struct {
	RWC      io.ReadWriteCloser
	Protocol protocol.Protocol
	KeepLive time.Duration
	Ctx      context.Context
	Flag     flagEnum
}

type RawSession struct {
	id       string
	rwc      io.ReadWriteCloser
	delay    *ticker.Ticker
	cp       protocol.Protocol
	ctx      context.Context
	cl       context.CancelFunc
	closer   sync.Once
	launcher XLauncher
	monitor  xnetutil.Monitor
}

func NewSession(cfg SessionConfig) *RawSession {
	if cfg.Ctx == nil {
		cfg.Ctx = context.Background()
	}
	s := &RawSession{
		id:      uuid.NewId(1),
		rwc:     cfg.RWC,
		cp:      cfg.Protocol,
		monitor: xnetutil.NewMonitor(),
	}
	if s.cp == nil {
		s.cp = &jsonprotocol.JsonProtocol{}
	}
	s.launcher = NewXLauncher(cfg.RWC, cfg.Protocol, cfg.Flag)
	ctx, cancel := context.WithCancel(cfg.Ctx)
	s.ctx, s.cl = ctx, cancel
	ctxtool.GWaitFunc(s.ctx, func() {
		_ = s.Close()
	})
	s.delay = ticker.NewTicker(s.ctx)
	if cfg.KeepLive != 0 {
		go s.keepLive(cfg.KeepLive)
	}
	return s
}

func (rs *RawSession) Id() string {
	return rs.id
}

func (rs *RawSession) Close() error {
	rs.closer.Do(func() {
		rs.cl()
		rs.delay.Stop()
		rs.monitor.Dead()
	})
	return rs.rwc.Close()
}

func (rs *RawSession) Context() context.Context {
	return rs.ctx
}

func (rs *RawSession) ReadXMsg() (xMsg *XMsg, n int, err error) {
	xMsg = new(XMsg)
	for r := true; r; {
		xMsg, n, err = rs.launcher.ReadXMsg()
		if err != nil {
			return nil, 0, err
		}
		rs.monitor.AddCount(n, 0)
		r, err = rs.preHandle(xMsg)
		if err != nil {
			return nil, 0, err
		}
	}
	return xMsg, n, nil
}

func (rs *RawSession) SendXMsg(header string, id uint32, opt OptType, data any) (xid uint32, n int, err error) {
	xid, n, err = rs.launcher.SendXMsg(header, id, opt, data)
	rs.monitor.AddCount(0, n)
	return xid, n, err
}

func (rs *RawSession) RecvXMsg(header string, id uint32, opt OptType, data any) (xid uint32, n int, err error) {
	xid, n, err = rs.launcher.RecvXMsg(header, id, opt, data)
	rs.monitor.AddCount(0, n)
	return xid, n, err
}

func (rs *RawSession) GetXMsgId() uint32 {
	i := rs.launcher.(*xLauncher)
	return i.XWriteLauncher.(*xWriteLauncher).getId()
}

func (rs *RawSession) GetDelay() time.Duration {
	return rs.monitor.GetDelay()
}

// Delay
//
// The call will get the latency of the session. Obviously,
// if the other party does not read the packet ReadCMsg, it will not get the latest latency data
//
// Output -1 if startup fails and 0 if ctx shuts down or times out
func (rs *RawSession) Delay(ctx context.Context) time.Duration {
	return rs.delay.DelayOnce(ctx, func(id string) error {
		_, _, err := rs.SendXMsg("", 0, optPing, id)
		return err
	})
}

// DelayTick
//
//	The same as that of the Delay, but will continue to output, encounters an error end will shut down the channel or context
func (rs *RawSession) DelayTick(ctx context.Context, interval time.Duration) <-chan time.Duration {
	return rs.delay.DelayTick(ctx, interval, func(id string) error {
		_, _, err := rs.SendXMsg("", 0, optPing, id)
		return err
	})
}

func (rs *RawSession) GetCount() (r, w uint64) {
	return rs.monitor.GetCount()
}

func (rs *RawSession) GetCountView() (r, w string) {
	return rs.monitor.GetCountView()
}

// Speed KeepLive needs to be set to true, otherwise the value will always be 0.
func (rs *RawSession) Speed() (r, w float64) {
	return rs.monitor.Speed()
}

// SpeedView KeepLive needs to be set to true, otherwise the value will always be 0.
func (rs *RawSession) SpeedView() (r, w string) {
	return rs.monitor.SpeedView()
}

// LifeDuration The duration from the creation of the object to the closure is recorded, not the duration from the time it starts working to the time it stops working.
func (rs *RawSession) LifeDuration() time.Duration {
	return rs.monitor.LifeDuration()
}

func (rs *RawSession) CreateTime() time.Time {
	return rs.monitor.CreateTime()
}

func (rs *RawSession) DeadTime() time.Time {
	return rs.monitor.DeadTime()
}

func (rs *RawSession) MonitorInfo() xnetutil.MonitorInfo {
	return rs.monitor.Info()
}

func (rs *RawSession) keepLive(d time.Duration) {
	tick := rs.DelayTick(context.Background(), d)
	for range tick {
		rs.monitor.RecordSpeed()
		rs.monitor.RecordDelay(rs.delay.GetDelay())
	}
}

func (rs *RawSession) preHandle(xMsg *XMsg) (r bool, err error) {
	switch xMsg.Opt() {
	case optPing:
		r = true
		id := ""
		err = xMsg.Unmarshal(&id)
		if err != nil {
			return r, err
		}
		_, _, err = rs.RecvXMsg("", xMsg.Id(), optPong, id)
	case optPong:
		r = true
		id := ""
		err = xMsg.Unmarshal(&id)
		if err != nil {
			return r, err
		}
		rs.delay.Record(id)
	}
	return r, err
}
