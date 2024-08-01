package multiplex

import (
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const buffSize = 1024

type writeFunc func([]byte) error

func newSession(ctx context.Context, cr func(), wl writeFunc) *mSession {
	sess := &mSession{
		wl: wl,
		cr: cr,
	}
	sess.ctx, sess.cancel = context.WithCancel(ctx)
	sess.read, sess.readCh = io.Pipe()
	return sess
}

type mSession struct {
	ctx    context.Context
	cancel context.CancelFunc
	read   *io.PipeReader
	readCh *io.PipeWriter
	wl     writeFunc
	once   sync.Once
	cr     func()

	activeTime atomic.Pointer[time.Time]
}

func (ms *mSession) LocalAddr() net.Addr {
	return nil
}

func (ms *mSession) RemoteAddr() net.Addr {
	return nil
}

func (ms *mSession) SetDeadline(t time.Time) error {
	return nil
}

func (ms *mSession) SetReadDeadline(t time.Time) error {
	return nil
}

func (ms *mSession) SetWriteDeadline(t time.Time) error {
	return nil
}

func (ms *mSession) Read(b []byte) (n int, err error) {
	err = ms.disable()
	if err != nil {
		return 0, err
	}
	return ms.read.Read(b)
}

func (ms *mSession) Write(b []byte) (n int, err error) {
	err = ms.disable()
	if err != nil {
		return 0, err
	}
	err = ms.wl(b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (ms *mSession) Close() error {
	ms.once.Do(func() {
		ms.cancel()
		ms.read.Close()
		ms.readCh.Close()
		ms.cr()
	})
	return ms.ctx.Err()
}

func (ms *mSession) disable() error {
	if ms.ctx.Err() != nil {
		return ms.Close()
	}
	return nil
}

func (ms *mSession) writeCh(data []byte) (err error) {
	ms.ping()
	_, err = ms.readCh.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (ms *mSession) ping() {
	t := time.Now()
	ms.activeTime.Store(&t)
}
