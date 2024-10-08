package xrpc

import (
	"fmt"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
)

type CryptoConfig struct {
	Name     string
	Crypto   pcrypto.PCrypto
	Priority int8
}

func (cc *CryptoConfig) String() string {
	return fmt.Sprintf("%s_%s", cc.Crypto.Name(), cc.Name)
}
