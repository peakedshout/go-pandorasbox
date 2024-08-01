package task

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestNewTaskCtx(t *testing.T) {
	taskCtx := NewTaskCtx[string](context.Background())
	go func() {
		time.Sleep(1 * time.Second)
		s := "hello,world!"
		taskCtx.CallBack(0, &s, nil)
	}()

	msg, err := taskCtx.RegisterTaskWithDr(2*time.Second, 0, func() error {
		log.Println("task running")
		return nil
	})
	log.Println("err:", err)
	log.Println("msg:", *msg)
}
