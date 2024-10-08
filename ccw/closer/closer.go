package closer

import (
	"errors"
	"sync"
)

type Closer interface {
	AddCloseFn(fn func())
	CloseErr(err error) error
	Close() error
	Err() error
}

var ErrIsClosed = errors.New("closer is closed")

type closer struct {
	errLock     sync.RWMutex
	err         error
	closer      sync.Once
	lock        sync.Mutex
	closed      bool
	closeFnList []func()
}

func NewCloser() Closer {
	return &closer{}
}

func (c *closer) AddCloseFn(fn func()) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		fn()
		return
	}
	c.closeFnList = append(c.closeFnList, fn)
}

func (c *closer) CloseErr(err error) error {
	c.closer.Do(func() {
		c.setErr(err)
		c.lock.Lock()
		defer c.lock.Unlock()
		c.closed = true
		for _, one := range c.closeFnList {
			one()
		}
	})
	return c.Err()
}

func (c *closer) Close() error {
	return c.CloseErr(ErrIsClosed)
}

func (c *closer) Err() error {
	c.errLock.RLock()
	defer c.errLock.RUnlock()
	return c.err
}

func (c *closer) setErr(err error) {
	c.errLock.Lock()
	defer c.errLock.Unlock()
	c.err = err
}
