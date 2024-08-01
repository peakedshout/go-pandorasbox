package expired

import (
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"time"
)

func NewScratchpads[T1 any, T2 any](e *ExpiredCtx, m *tmap.SyncMap[T1, T2], expiredTime time.Duration) *Scratchpads[T1, T2] {
	return &Scratchpads[T1, T2]{e: e, m: m, expiredTime: expiredTime}
}

type Scratchpads[T1 any, T2 any] struct {
	m           *tmap.SyncMap[T1, T2]
	e           *ExpiredCtx
	expiredTime time.Duration
}

type scratchpadsUnit struct {
	id  any
	eFn func()
}

func (s *scratchpadsUnit) Id() any {
	return s.id
}

func (s *scratchpadsUnit) ExpiredFunc() {
	s.eFn()
}

func (s *Scratchpads[T1, T2]) Set(k T1, v T2) {
	s.m.Store(k, v)
	s.e.SetWithDuration(s.newUnit(k), s.expiredTime)
}

func (s *Scratchpads[T1, T2]) newUnit(k T1) *scratchpadsUnit {
	return &scratchpadsUnit{
		id: k,
		eFn: func() {
			s.m.Delete(k)
		},
	}
}
