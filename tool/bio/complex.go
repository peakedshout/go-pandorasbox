package bio

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/ccw/closer"
	"io"
	"sync"
	"sync/atomic"
)

type Broadcast struct {
	listener  sync.Map
	spokesman chan []byte
	count     int64
	ctx       context.Context
	cancel    context.CancelFunc
	closer    closer.Closer
}

func NewBroadcastWithReader(r io.Reader, size int) *Broadcast {
	bc := NewBroadcast()
	go func() {
		for {
			buf := make([]byte, size)
			n, err := r.Read(buf)
			if err != nil {
				bc.closer.CloseErr(err)
				return
			}
			select {
			case bc.spokesman <- buf[:n]:
			case <-bc.ctx.Done():
				return
			}
		}
	}()
	return bc
}

func NewBroadcast() *Broadcast {
	bc := &Broadcast{
		spokesman: make(chan []byte),
		count:     0,
		closer:    closer.NewCloser(),
	}
	bc.ctx, bc.cancel = context.WithCancel(context.Background())
	bc.closer.AddCloseFn(func() {
		bc.cancel()
	})

	go func() {
		for {
			select {
			case data := <-bc.spokesman:
				bc.allRWCopy(data)
			case <-bc.ctx.Done():
				return
			}
		}
	}()
	return bc
}

func (bc *Broadcast) Write(b []byte) (n int, err error) {
	select {
	case bc.spokesman <- b:
		return len(b), nil
	case <-bc.ctx.Done():
		return 0, nil
	}
}

func (bc *Broadcast) SetListener(wc io.WriteCloser) int64 {
	id := atomic.AddInt64(&bc.count, 1)
	bc.listener.Store(id, wc)
	return id
}

func (bc *Broadcast) DelListener(id int64) {
	bc.listener.Delete(id)
}

func (bc *Broadcast) DelAndCloseListener(id int64) {
	a, ok := bc.listener.LoadAndDelete(id)
	if ok {
		a.(io.WriteCloser).Close()
	}
}

func (bc *Broadcast) allRWCopy(b []byte) {
	bc.listener.Range(func(key, value any) bool {
		if _, err := value.(io.WriteCloser).Write(b); err != nil {
			bc.listener.Delete(key)
			value.(io.WriteCloser).Close()
		}
		return true
	})
}
