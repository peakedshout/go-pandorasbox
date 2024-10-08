package icrypto

import "github.com/peakedshout/go-pandorasbox/uerror"

var (
	ErrCheckKeyLen       = uerror.NewErrorCode(3100, 1000, "unexpected key len: %d must be %v")
	ErrEncrypt           = uerror.NewErrorCode(3100, 1001, "encrypt: <%v>")
	ErrDecrypt           = uerror.NewErrorCode(3100, 1002, "decrypt: <%v>")
	ErrNotSupportType    = uerror.NewErrorCode(3100, 1003, "not support type: %v")
	ErrUnexpectedKeys    = uerror.NewErrorCode(3100, 1004, "unexpected number of keys: %d must be %d")
	ErrPublicKeyInvalid  = uerror.NewErrorCode(3100, 1005, "publicKey invalid: %v")
	ErrPrivateKeyInvalid = uerror.NewErrorCode(3100, 1006, "privateKey invalid: %v")
)
