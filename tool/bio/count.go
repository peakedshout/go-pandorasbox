package bio

import (
	"fmt"
	"sync/atomic"
)

func NewIOCounter(r, w uint64) *IOCounter {
	return &IOCounter{
		r: NewCounter(r),
		w: NewCounter(w),
	}
}

type IOCounter struct {
	r *Counter
	w *Counter
}

func (c *IOCounter) Add(r, w any) {
	c.r.Add(r)
	c.w.Add(w)
}

func (c *IOCounter) Get() (r, w uint64) {
	return c.r.Get(), c.w.Get()
}

func (c *IOCounter) GetView() (r, w string) {
	return c.r.GetView(), c.w.GetView()
}

func (c *IOCounter) Set(r, w any) {
	c.r.Set(r)
	c.w.Set(w)
}

func NewCounter(i uint64) *Counter {
	return &Counter{i: i}
}

type Counter struct {
	i uint64
}

func (c *Counter) Add(a any) {
	switch i := a.(type) {
	case int:
		if i == 0 {
			return
		}
		atomic.AddUint64(&c.i, uint64(i))
	case uint:
		if i == 0 {
			return
		}
		atomic.AddUint64(&c.i, uint64(i))
	case int8:
		if i == 0 {
			return
		}
		atomic.AddUint64(&c.i, uint64(i))
	case uint8:
		if i == 0 {
			return
		}
		atomic.AddUint64(&c.i, uint64(i))
	case int16:
		if i == 0 {
			return
		}
		atomic.AddUint64(&c.i, uint64(i))
	case uint16:
		if i == 0 {
			return
		}
		atomic.AddUint64(&c.i, uint64(i))
	case int32:
		if i == 0 {
			return
		}
		atomic.AddUint64(&c.i, uint64(i))
	case uint32:
		if i == 0 {
			return
		}
		atomic.AddUint64(&c.i, uint64(i))
	case int64:
		if i == 0 {
			return
		}
		atomic.AddUint64(&c.i, uint64(i))
	case uint64:
		if i == 0 {
			return
		}
		atomic.AddUint64(&c.i, uint64(i))
	default:
		panic("invalid int type")
	}
}

func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.i)
}

func (c *Counter) GetView() string {
	return formatCount(c.Get())
}

func (c *Counter) Set(a any) {
	switch i := a.(type) {
	case int:
		atomic.StoreUint64(&c.i, uint64(i))
	case uint:
		atomic.StoreUint64(&c.i, uint64(i))
	case int8:
		atomic.StoreUint64(&c.i, uint64(i))
	case uint8:
		atomic.StoreUint64(&c.i, uint64(i))
	case int16:
		atomic.StoreUint64(&c.i, uint64(i))
	case uint16:
		atomic.StoreUint64(&c.i, uint64(i))
	case int32:
		atomic.StoreUint64(&c.i, uint64(i))
	case uint32:
		atomic.StoreUint64(&c.i, uint64(i))
	case int64:
		atomic.StoreUint64(&c.i, uint64(i))
	case uint64:
		atomic.StoreUint64(&c.i, uint64(i))
	default:
		panic("invalid int type")
	}
}

func formatCount(count uint64) string {
	f := float64(count)
	if f < 1024 {
		return fmt.Sprintf("%.2fB", f/float64(1))
	} else if count < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", f/float64(1024))
	} else if count < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", f/float64(1024*1024))
	} else if f < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", f/float64(1024*1024*1024))
	} else if f < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", f/float64(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fEB", f/float64(1024*1024*1024*1024*1024))
	}
}

func FormatCount(count uint64) string {
	return formatCount(count)
}
