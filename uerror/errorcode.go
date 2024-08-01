package uerror

import (
	"fmt"
)

type ErrorCode struct {
	errType int
	errCode int
	msg     string
}

func NewErrorCode(t int, c int, m string) *ErrorCode {
	ec := &ErrorCode{
		errType: t,
		errCode: c,
		msg:     m,
	}
	if _muxDebug {
		_, ok := _cacheMap[ec.Code()]
		if ok {
			panic(ec.Code() + " already exists")
		}
	}
	_cacheMap[ec.Code()] = ec
	return ec
}

func (ec *ErrorCode) Code() string {
	return fmt.Sprintf("%d_%d", ec.errType, ec.errCode)
}

func (ec *ErrorCode) Errorf(a ...any) *UError {
	ue := &UError{
		ec: ec,
	}
	ue.WithParamList(a...)
	return ue
}
