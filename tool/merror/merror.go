package merror

import (
	"errors"
	"fmt"
	"sync"
)

func NewMultiErr(pre string) *MultiErr {
	return &MultiErr{
		pre:  pre,
		mux:  sync.Mutex{},
		errs: make(map[string]error),
	}
}

type MultiErr struct {
	pre  string
	mux  sync.Mutex
	errs map[string]error
}

func (e *MultiErr) AddErr(code string, err error) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.errs[code] = err
}

func (e *MultiErr) Error() string {
	e.mux.Lock()
	defer e.mux.Unlock()
	b := []byte(fmt.Sprintf("multiErr %s:", e.pre))
	for i, err := range e.errs {
		b = append(b, '\n')
		b = append(b, fmt.Errorf("\t%s: %w", i, err).Error()...)
	}
	return string(b)
}

func (e *MultiErr) Is(err error) bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	for _, err2 := range e.errs {
		if errors.Is(err2, err) {
			return true
		}
	}
	return false
}

func (e *MultiErr) As(a any) bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	for _, err2 := range e.errs {
		if errors.As(err2, &a) {
			return true
		}
	}
	return false
}

func (e *MultiErr) Nil() bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	return len(e.errs) == 0
}
