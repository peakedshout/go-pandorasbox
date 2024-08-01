package ticker

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTicker(t *testing.T) {
	tr := NewTicker(nil)
	defer tr.Stop()
	did := ""
	go func() {
		time.Sleep(1 * time.Second)
		tr.Record(did)
	}()
	to := tr.DelayOnce(nil, func(id string) error {
		did = id
		return nil
	})
	fmt.Println(to)
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(1 * time.Second)
			tr.Record(did)
		}
		time.Sleep(1 * time.Second)
		tr.Stop()
	}()
	ch := tr.DelayTick(nil, 0*time.Second, func(id string) error {
		did = id
		return nil
	})
	for one := range ch {
		fmt.Println(one)
	}
}
