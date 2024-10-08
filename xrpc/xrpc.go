package xrpc

import (
	"context"
	"time"
)

type Rpc interface {
	Bind(out any) error
	Context() context.Context
}

type Stream interface {
	Id() string
	Recv(out any) error
	Send(data any) error
	Close() error
	Context() context.Context
	Type() string
	GetCount() (r uint64, w uint64)
	GetCountView() (r, w string)
	Speed() (r float64, w float64)
	SpeedView() (r string, w string)
	LifeDuration() time.Duration
	CreateTime() time.Time
	DeadTime() time.Time
	GetDelay() time.Duration
}

type SendStream interface {
	Send(data any) error
	Bind(out any) error
	Close() error
	Context() context.Context
}

type RecvStream interface {
	Recv(out any) error
	Close() error
	Context() context.Context
}

type ReverseRpc interface {
	Bind(out any) error
	Context() context.Context
	Rpc(ctx context.Context, header string, send, recv any) error
}
