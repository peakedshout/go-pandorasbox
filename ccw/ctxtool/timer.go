package ctxtool

import (
	"context"
	"time"
)

func RunTimerFunc(ctx context.Context, td time.Duration, fn func(ctx context.Context) error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := fn(ctx); err != nil {
		return err
	}
	timer := time.NewTimer(td)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := fn(ctx); err != nil {
				return err
			}
			timer.Reset(td)
		}
	}
}
