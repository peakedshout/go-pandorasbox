package speed

import (
	"fmt"
	"sync"
	"time"
)

type Ticker struct {
	read  int
	write int
	rLock sync.Mutex
	wLock sync.Mutex
	t     time.Time
}

func (st *Ticker) Set(i int) {
	if i == 0 {
		return
	}
	st.wLock.Lock()
	defer st.wLock.Unlock()
	t := time.Now().Round(time.Second)
	if !st.t.Equal(t) {
		st.rLock.Lock()
		st.read = st.write
		st.write = i
		st.t = t
		st.rLock.Unlock()
	} else {
		st.write += i
	}
}

func (st *Ticker) Get() int {
	st.rLock.Lock()
	defer st.rLock.Unlock()
	t := time.Now().Round(time.Second)
	s := t.Sub(st.t)
	if s > time.Second {
		if s <= 2*time.Second {
			st.wLock.Lock()
			st.read = st.write
			st.write = 0
			st.t = t
			st.wLock.Unlock()
		}
		st.read = 0
	}
	return st.read
}

func formatSpeed(speed int) (size string) {
	fileSize := int64(speed)
	if fileSize < 1024 {
		return fmt.Sprintf("%.2fB/s", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB/s", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB/s", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB/s", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB/s", float64(fileSize)/float64(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fEB/s", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}

func FormatSpeed(speed int) string {
	return formatSpeed(speed)
}
