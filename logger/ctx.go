package logger

import "context"

const _ctxLogKey = "ctx_log_key"

var _zeroLogger Logger

func init() {
	_zeroLogger = Init("zero")
	_zeroLogger.SetLoggerLevel(LogLevelOff)
}

func MustLogger(ctx context.Context) Logger {
	l := GetLogger(ctx)
	if l == nil {
		return _zeroLogger
	}
	return l
}

func GetLogger(ctx context.Context) Logger {
	value := ctx.Value(_ctxLogKey)
	if value == nil {
		return nil
	}
	l, _ := value.(Logger)
	return l
}

func SetLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, _ctxLogKey, l)
}
