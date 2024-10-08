package ctxtool

import (
	"context"
	"io"
	"time"
)

func NewRcContext(ctx context.Context, rc io.ReadCloser) context.Context {
	nCtx, _ := monitorConn(ctx, rc)
	return nCtx
}

func NewRcContextWithCancel(ctx context.Context, rc io.ReadCloser) (context.Context, context.CancelFunc) {
	return monitorConn(ctx, rc)
}

func monitorConn(ctx context.Context, rc io.ReadCloser) (context.Context, context.CancelFunc) {
	zero := make([]byte, 0)
	tr := time.NewTimer(0 * time.Second)
	if ctx == nil {
		ctx = context.Background()
	}
	nCtx, cl := context.WithCancel(ctx)
	go func() {
		defer tr.Stop()
		defer cl()
		for {
			_, err := rc.Read(zero)
			if err != nil {
				cl()
				return
			}
			if !tr.Stop() {
				<-tr.C
			}
			tr.Reset(1 * time.Second)
			select {
			case <-nCtx.Done():
				if ctx.Err() != nil {
					_ = rc.Close()
				}
				return
			case <-tr.C:
			}
		}
	}()
	return nCtx, cl
}

func RcTry(rc io.ReadCloser) error {
	_, err := rc.Read([]byte{})
	if err != nil {
		return err
	}
	return nil
}
