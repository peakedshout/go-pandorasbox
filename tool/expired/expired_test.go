package expired

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func getD() time.Duration {

	return time.Duration(rand.Intn(1*1000)+2*1000) * time.Millisecond

}

type tt struct {
	a  int
	tk time.Time
	wg *sync.WaitGroup
}

func (t *tt) Id() any {
	return t.a
}

func (t *tt) ExpiredFunc() {
	if time.Now().Sub(t.tk) > 1000*time.Millisecond {
		panic("time out " + time.Now().Sub(t.tk).String())
	}
	if time.Now().Before(t.tk) {
		panic("????")
	}
	t.wg.Done()
}

func Test10(t *testing.T) {
	i := 0
	defer func() {
		fmt.Println("num:", i)
	}()
	var wg sync.WaitGroup
	for i = 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
			}()
			TestHeap(t)
		}()
	}
	wg.Wait()
}

func TestUpdate(t *testing.T) {
	e := Init(nil, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	t1 := &tt{
		a:  1,
		tk: time.Now().Add(2 * time.Second),
		wg: &wg,
	}
	e.SetWithTime(t1, t1.tk)
	t1.tk = time.Now().Add(5 * time.Second)
	e.UpdateWithTime(1, t1.tk)
	wg.Wait()
}

func TestHeap(t *testing.T) {
	e := Init(nil, 100)
	tn := time.Now()
	n := 0
	var d time.Time

	var ll []tt
	var wg sync.WaitGroup

	for i := 0; i < 10000*10; i++ {
		wg.Add(1)
		n = rand.Intn(10000)
		d = tn.Add(getD())
		tt1 := tt{
			a:  n,
			tk: d,
			wg: &wg,
		}
		ll = append(ll, tt1)
		e.SetWithTime(&tt1, d)
	}
	wg.Wait()
	//time.Sleep(5 * time.Second)
	//e.Wait()
}

func BenchmarkExpiredCtx(b *testing.B) {
	b.N = rand.Intn(100)
	for i := 0; i < b.N; i++ {
		b.Run("", BenchmarkExpired)
	}
}

func BenchmarkExpired(b *testing.B) {
	e := Init(nil, 100)
	defer e.Stop()
	tn := time.Now()
	n := 0
	var d time.Time

	var ll []tt
	var wg sync.WaitGroup

	for i := 0; i < 10000*10; i++ {
		wg.Add(1)
		n = rand.Intn(10000)
		d = tn.Add(getD())
		tt1 := tt{
			a:  n,
			tk: d,
			wg: &wg,
		}
		ll = append(ll, tt1)
		e.SetWithTime(&tt1, d)
	}
	wg.Wait()
}
