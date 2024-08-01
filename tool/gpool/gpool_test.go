package gpool

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewGPool(t *testing.T) {
	gp := NewGPool(context.Background(), 100)
	gp.Do(func() {
		fmt.Println("???")
	})
	time.Sleep(1 * time.Second)
	gp.Wait()
	gp.Stop()
	gp.Wait()
}

func TestEmpty(t *testing.T) {
	gp := NewGPool(context.Background(), 100)
	gp.Do(func() {
		fmt.Println("???")
	})
	time.Sleep(1 * time.Second)
	gp.Wait()
	gp.Do(func() {
		fmt.Println("???")
	})
	gp.Do(func() {
		fmt.Println("???")
	})
	time.Sleep(1 * time.Second)
	gp.Wait()
	gp.Do(func() {
		fmt.Println("???")
	})
	gp.Do(func() {
		fmt.Println("???")
	})
	gp.Stop()
	gp.Wait()
}
