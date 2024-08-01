package xmsg

import (
	"github.com/peakedshout/go-pandorasbox/tool/xerror"
)

var (
	ErrHeaderMustBeLessEqual     = xerror.New("marshal: header must be less than or equal to 255")
	ErrDataOutputToNonNilPointer = xerror.New("unmarshal: %s non-nil pointer")
	ErrDataOutputTypeInvalid     = xerror.New("unmarshal: output type invalid")
	ErrDataOutputNotData         = xerror.New("unmarshal: not data")
	ErrDataOutputError           = xerror.New("%v")
)
