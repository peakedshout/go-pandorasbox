package xneterr

import "github.com/peakedshout/go-pandorasbox/uerror"

var (
	ErrConnTypeIsInvalid = uerror.NewErrorCode(7100, 1000, "conn type invalid: conn is not %v")
	ErrNetworkIsInvalid  = uerror.NewErrorCode(7100, 1001, "network invalid: %v")
	ErrNilTlsConfig      = uerror.NewErrorCode(7100, 1002, "%v nil tls config")
)
