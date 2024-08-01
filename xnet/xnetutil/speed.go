package xnetutil

import (
	"fmt"
	"sync"
	"time"
)

func FormatSpeed(speed float64) string {
	return formatSpeed(speed)
}

func formatSpeed(speed float64) string {
	if speed < 1024 {
		return fmt.Sprintf("%.2fB/s", speed/float64(1))
	} else if speed < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB/s", speed/float64(1024))
	} else if speed < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB/s", speed/float64(1024*1024))
	} else if speed < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB/s", speed/float64(1024*1024*1024))
	} else if speed < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB/s", speed/float64(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fEB/s", speed/float64(1024*1024*1024*1024*1024))
	}
}

// NewSpeedometer num must be > 0
func NewSpeedometer(num int) *Speedometer {
	if num <= 0 {
		panic("invalid num")
	}
	return &Speedometer{
		speed: make([]float64, num),
		l:     num,
	}
}

type Speedometer struct {
	mux   sync.Mutex
	lastT time.Time
	lastV []uint64
	speed []float64
	once  bool
	l     int
}

func (s *Speedometer) Set(vs ...uint64) {
	if s.l != len(vs) {
		panic("the number of objects is inconsistent")
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	t := time.Now()
	if !s.once {
		s.lastV = vs
		s.lastT = t
		s.once = true
		return
	}
	td := t.Sub(s.lastT)
	if td == 0 {
		td = 500
	}
	for i, v := range vs {
		if v == s.lastV[i] {
			s.speed[i] = 0
		} else if v < s.lastV[i] {
			s.lastV[i] = v
			s.speed[i] = 0
		} else {
			s.speed[i] = float64(v-s.lastV[i]) / td.Seconds()
			s.lastV[i] = v
		}
	}
	s.lastT = t
}

func (s *Speedometer) Speed() []float64 {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.speed
}

func (s *Speedometer) View() []string {
	speed := s.Speed()
	sl := make([]string, 0, len(speed))
	for _, f := range speed {
		sl = append(sl, formatSpeed(f))
	}
	return sl
}
