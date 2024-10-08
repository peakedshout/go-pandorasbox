package expired

import (
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"time"
)

func NewTODO(e *ExpiredCtx) *TODO {
	return &TODO{e: e}
}

type TODO struct {
	e *ExpiredCtx
}

type todoUnit struct {
	id  any
	eFn func()
}

func (t *todoUnit) Id() any {
	return t.id
}

func (t *todoUnit) ExpiredFunc() {
	t.eFn()
}

func (t *TODO) Duration(td time.Duration, fn func()) bool {
	return t.e.SetWithDuration(t.newUnit(fn), td)
}

func (t *TODO) Time(tt time.Time, fn func()) bool {
	return t.e.SetWithTime(t.newUnit(fn), tt)
}

func (t *TODO) newUnit(fn func()) *todoUnit {
	return &todoUnit{
		id:  uuid.NewId(1),
		eFn: fn,
	}
}
