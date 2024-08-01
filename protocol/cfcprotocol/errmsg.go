package cfcprotocol

import "github.com/peakedshout/go-pandorasbox/uerror"

var (
	errCFCProtocolWaitPacket = uerror.NewErrorCode(3200, 1000, "cfc protocol: wait pack")

	//ErrCFCProtocolSkipParse       = uerror.NewErrorCode(3200, 1001, "cfc protocol: skip parse")

	ErrCFCProtocolIsNotGoCFC      = uerror.NewErrorCode(3200, 1002, "cfc protocol version: %s must be %s")
	ErrCFCProtocolLensTooShort    = uerror.NewErrorCode(3200, 1003, "cfc protocol lens: %d too small to %d bytes")
	ErrCFCProtocolLensTooLong     = uerror.NewErrorCode(3200, 1004, "cfc protocol lens: %d too long to %d bytes")
	ErrCFCProtocolHashCheckFailed = uerror.NewErrorCode(3200, 1005, "cfc protocol hash: check failed")

	ErrCFCProtocolDecodeToNonNilPointer  = uerror.NewErrorCode(3200, 1020, "cfc protocol decode: %s be non-nil pointer")
	ErrCFCProtocolDecodeNilData          = uerror.NewErrorCode(3200, 1021, "cfc protocol decode: nil data")
	ErrCFCProtocolDecodeInvalidMsgType   = uerror.NewErrorCode(3200, 1022, "cfc protocol decode: invalid msg type")
	ErrCFCProtocolDecodeInvalidContainer = uerror.NewErrorCode(3200, 1023, "cfc protocol decode: invalid container")
)
