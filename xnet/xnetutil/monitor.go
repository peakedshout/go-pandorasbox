package xnetutil

import (
	"github.com/peakedshout/go-pandorasbox/tool/bio"
	"sync/atomic"
	"time"
)

type Monitor interface {
	AddCount(r any, w any)
	GetCount() (r uint64, w uint64)
	GetCountView() (r string, w string)
	RecordSpeed()
	Speed() (r float64, w float64)
	SpeedView() (r string, w string)
	Dead()
	LifeDuration() time.Duration
	CreateTime() time.Time
	DeadTime() time.Time
	RecordDelay(td time.Duration)
	GetDelay() time.Duration
	Info() MonitorInfo
}

func NewMonitor() Monitor {
	m := &monitor{
		counter: bio.NewIOCounter(0, 0),
		speed:   NewSpeedometer(2),
		ct:      time.Now(),
	}
	m.delay.Store(new(time.Duration))
	return m
}

type monitor struct {
	counter *bio.IOCounter
	speed   *Speedometer
	ct      time.Time
	dt      atomic.Pointer[time.Time]
	delay   atomic.Pointer[time.Duration]
}

func (m *monitor) AddCount(r, w any) {
	m.counter.Add(r, w)
}

func (m *monitor) GetCount() (r uint64, w uint64) {
	return m.counter.Get()
}

func (m *monitor) GetCountView() (r string, w string) {
	return m.counter.GetView()
}

func (m *monitor) RecordSpeed() {
	m.speed.Set(m.counter.Get())
}

func (m *monitor) Speed() (r, w float64) {
	speed := m.speed.Speed()
	return speed[0], speed[1]
}

func (m *monitor) SpeedView() (r, w string) {
	view := m.speed.View()
	return view[0], view[1]
}

func (m *monitor) Dead() {
	t := time.Now()
	m.dt.CompareAndSwap(nil, &t)
}

func (m *monitor) LifeDuration() time.Duration {
	t := m.dt.Load()
	if t == nil {
		return time.Since(m.ct)
	} else {
		return t.Sub(m.ct)
	}
}

func (m *monitor) CreateTime() time.Time {
	return m.ct
}

func (m *monitor) DeadTime() time.Time {
	t := m.dt.Load()
	if t == nil {
		return time.Time{}
	}
	return *t
}

func (m *monitor) RecordDelay(td time.Duration) {
	m.delay.Store(&td)
}

func (m *monitor) GetDelay() time.Duration {
	return *m.delay.Load()
}

func (m *monitor) Info() MonitorInfo {
	return FormatMonitorInfo(m)
}

type MonitorInfo struct {
	RCount, WCount         uint64
	RCountView, WCountView string
	RSpeed, WSpeed         float64
	RSpeedView, WSpeedView string
	CreateTime, DeadTime   time.Time
	LifeDuration           time.Duration
	Delay                  time.Duration
}

// FormatMonitorInfo
// RCount, WCount RCountView, WCountView GetCount() (r uint64, w uint64)
// RSpeed, WSpeed RSpeedView, WSpeedView Speed() (r float64, w float64)
// CreateTime CreateTime() time.Time
// DeadTime DeadTime() time.Time
// LifeDuration CreateTime() time.Time DeadTime() time.Time
// Delay GetDelay() time.Duration
func FormatMonitorInfo(a any) MonitorInfo {
	info := MonitorInfo{}
	if i, ok := a.(interface{ GetDelay() time.Duration }); ok {
		info.Delay = i.GetDelay()
	}
	if i, ok := a.(interface{ CreateTime() time.Time }); ok {
		info.CreateTime = i.CreateTime()
	}
	if i, ok := a.(interface{ DeadTime() time.Time }); ok {
		info.DeadTime = i.DeadTime()
		if !info.CreateTime.IsZero() {
			if !info.DeadTime.IsZero() {
				info.LifeDuration = info.DeadTime.Sub(info.CreateTime)
			} else {
				info.LifeDuration = time.Since(info.CreateTime)
			}
		}
	}
	if i, ok := a.(interface{ GetCount() (r uint64, w uint64) }); ok {
		info.RCount, info.WCount = i.GetCount()
		info.RCountView, info.WCountView = bio.FormatCount(info.RCount), bio.FormatCount(info.WCount)
	}
	if i, ok := a.(interface{ Speed() (r float64, w float64) }); ok {
		info.RSpeed, info.WSpeed = i.Speed()
		info.RSpeedView, info.WSpeedView = FormatSpeed(info.RSpeed), FormatSpeed(info.WSpeed)
	}
	return info
}
