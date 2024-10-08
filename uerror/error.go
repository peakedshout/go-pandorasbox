package uerror

import (
	"errors"
	"fmt"
	"strings"
)

var _muxDebug = true
var _cacheMap = make(map[string]*ErrorCode)

type UError struct {
	ec        *ErrorCode
	paramList []any
	wrapError error

	_reverse bool
}

func (ue *UError) Code() string {
	return ue.ec.Code()
}

func (ue *UError) Error() string {
	return fmt.Sprintf("[%s]{%s}", ue.Code(), ue.String())
}

func (ue *UError) String() string {
	if ue._reverse {
		return fmt.Sprint(ue.paramList...)
	}
	return fmt.Sprintf(ue.ec.msg, ue.paramList...)
}

func (ue *UError) WithParamList(a ...any) {
	for _, one := range a {
		if err, ok := one.(error); ok {
			ue.wrapError = err
			break
		}
	}
	ue.paramList = a
}

func (ue *UError) Wrap(err error) {
	ue.wrapError = err
}

func (ue *UError) Unwrap() error {
	return ue.wrapError
}

func (ue *UError) Is(target error) bool {
	tue, ok := target.(*UError)
	if !ok {
		return false
	}
	return ue.Code() == tue.Code()
}

func (ue *UError) IsErrorCode(code *ErrorCode) bool {
	return Is(ue, code)
}

func Is(err error, code *ErrorCode) bool {
	var ue *UError
	for errors.As(err, &ue) {
		if ue.Code() == code.Code() {
			return true
		}
		if ue.wrapError != nil {
			err = ue.wrapError
		} else {
			break
		}
	}
	return false
}

func Unmarshal(str string) (code, msg string, ok bool) {
	if len(str) == 0 || str[0] != '[' || str[len(str)-1] != '}' {
		return "", "", false
	}
	str = str[1 : len(str)-1]
	sl := strings.SplitN(str, "]{", 2)
	if len(sl) != 2 {
		return "", "", false
	}
	return sl[0], sl[1], true
}

func unmarshalInside(str string) (sl []string, ok bool) {
	i := strings.Index(str, "[")
	if i == -1 {
		return nil, false
	}
	j := strings.LastIndex(str, "}")
	if j == -1 {
		return nil, false
	}
	sl = []string{str[:i], str[i : j+1], str[j+1:]}
	return sl, true
}

func unmarshalUError(str string) ([]any, *UError, bool) {
	inside, ok := unmarshalInside(str)
	if !ok {
		return nil, nil, false
	}
	code, msg, ok := Unmarshal(inside[1])
	if !ok {
		return nil, nil, false
	}
	ec, ok := _cacheMap[code]
	if !ok {
		return nil, nil, false
	}
	rue := &UError{
		ec:       ec,
		_reverse: true,
	}
	al, ue, ok := unmarshalUError(msg)
	if ok {
		rue.WithParamList(al[0], ue, al[1])
	} else {
		rue.WithParamList(msg)
	}
	return []any{inside[0], inside[2]}, rue, true
}

func UnmarshalErrorCode(str string) (*ErrorCode, bool) {
	code, _, ok := Unmarshal(str)
	if !ok {
		return nil, false
	}
	ec, ok := _cacheMap[code]
	if !ok {
		return nil, false
	}
	return ec, true
}

func UnmarshalUError(str string) (*UError, bool) {
	_, ue, ok := unmarshalUError(str)
	if ok {
		return ue, true
	}
	return nil, false
}
