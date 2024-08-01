package xerror

import (
	"errors"
	"fmt"
)

func New(str string) *XError {
	return &XError{
		root: str,
		warp: errors.New(str),
	}
}

type XError struct {
	root    string
	warp    error
	display string
}

func (x *XError) Errorf(a ...any) error {
	return &XError{
		root: x.root,
		warp: fmt.Errorf(x.root, a...),
	}
}

func (x *XError) Error() string {
	return x.warp.Error()
}

func (x *XError) Is(target error) bool {
	if tx, ok := target.(*XError); ok {
		return tx.root == x.root
	}
	return target == x
}

func (x *XError) Unwrap() error {
	return x.warp
}

func (x *XError) As(a any) bool {
	switch e := a.(type) {
	case **XError:
		*e = x
		return true
	default:
		return false
	}
}
