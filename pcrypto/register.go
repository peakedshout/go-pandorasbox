package pcrypto

import (
	"github.com/peakedshout/go-pandorasbox/pcrypto/aescbc"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aescfb"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aescrt"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aesecb"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aesgcm"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aesofb"
	"github.com/peakedshout/go-pandorasbox/pcrypto/icrypto"
	"github.com/peakedshout/go-pandorasbox/pcrypto/rsa"
)

type Initiator struct{}

func (i Initiator) Name() string {
	return "pcrypto"
}

func (i Initiator) Init() error {
	RegisterAll()
	return nil
}

func RegisterAll() {
	icrypto.Register(aescbc.PCryptoAes128Cbc)
	icrypto.Register(aescbc.PCryptoAes192Cbc)
	icrypto.Register(aescbc.PCryptoAes256Cbc)
	icrypto.Register(aescbc.PCryptoAesHashCbc)
	icrypto.Register(aescbc.PCryptoAesXxxCbc)

	icrypto.Register(aescfb.PCryptoAes128Cfb)
	icrypto.Register(aescfb.PCryptoAes192Cfb)
	icrypto.Register(aescfb.PCryptoAes256Cfb)
	icrypto.Register(aescfb.PCryptoAesHashCfb)
	icrypto.Register(aescfb.PCryptoAesXxxCfb)

	icrypto.Register(aescrt.PCryptoAes128Crt)
	icrypto.Register(aescrt.PCryptoAes192Crt)
	icrypto.Register(aescrt.PCryptoAes256Crt)
	icrypto.Register(aescrt.PCryptoAesHashCrt)
	icrypto.Register(aescrt.PCryptoAesXxxCrt)

	icrypto.Register(aesecb.PCryptoAes128Ecb)
	icrypto.Register(aesecb.PCryptoAes192Ecb)
	icrypto.Register(aesecb.PCryptoAes256Ecb)
	icrypto.Register(aesecb.PCryptoAesHashEcb)
	icrypto.Register(aesecb.PCryptoAesXxxEcb)

	icrypto.Register(aesgcm.PCryptoAes128Gcm)
	icrypto.Register(aesgcm.PCryptoAes192Gcm)
	icrypto.Register(aesgcm.PCryptoAes256Gcm)
	icrypto.Register(aesgcm.PCryptoAesHashGcm)
	icrypto.Register(aesgcm.PCryptoAesXxxGcm)

	icrypto.Register(aesofb.PCryptoAes128Ofb)
	icrypto.Register(aesofb.PCryptoAes192Ofb)
	icrypto.Register(aesofb.PCryptoAes256Ofb)
	icrypto.Register(aesofb.PCryptoAesHashOfb)
	icrypto.Register(aesofb.PCryptoAesXxxOfb)

	icrypto.Register(rsa.PCryptoRsa)
	icrypto.Register(rsa.PCryptoRsaChunks)
	icrypto.Register(rsa.PCryptoRsaCert)
	icrypto.Register(rsa.PCryptoRsaCertChunks)
}

func GetCrypto(crypto string, keys ...[]byte) (PCrypto, error) {
	if crypto == CryptoPlaintext.Name() {
		return CryptoPlaintext, nil
	}
	i, err := icrypto.GetInterface(crypto)
	if err != nil {
		return nil, err
	}
	return NewPCrypto(i, keys...)
}

func IsSymmetric(crypto string) (bool, error) {
	if crypto == CryptoPlaintext.Name() {
		return CryptoPlaintext.IsSymmetric(), nil
	}
	i, err := icrypto.GetInterface(crypto)
	if err != nil {
		return false, err
	}
	return i.IsSymmetric(), nil
}

func GetAllCryptoNameList() []string {
	all := icrypto.GetAllInterface()
	list := make([]string, 0, len(all)+1)
	list = append(list, CryptoPlaintext.Name())
	for _, i := range all {
		list = append(list, i.Name())
	}
	return list
}
