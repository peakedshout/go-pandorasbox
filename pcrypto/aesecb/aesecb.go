package aesecb

import (
	"crypto/aes"
	"github.com/peakedshout/go-pandorasbox/pcrypto/icrypto"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
)

var (
	PCryptoAesHashEcb = AesEcb{name: "aes-hash-ecb", isHashKey: true}
	PCryptoAesXxxEcb  = AesEcb{name: "aes-xxx-ecb", keyLens: icrypto.AesKeyLens}
	PCryptoAes128Ecb  = AesEcb{name: "aes-128-ecb", keyLens: []int{16}}
	PCryptoAes192Ecb  = AesEcb{name: "aes-192-ecb", keyLens: []int{24}}
	PCryptoAes256Ecb  = AesEcb{name: "aes-256-ecb", keyLens: []int{32}}
)

type AesEcb struct {
	name      string
	keyLens   []int
	isHashKey bool
}

func (ae AesEcb) Name() string {
	return ae.name
}

func (ae AesEcb) IsSymmetric() bool {
	return true
}

func (ae AesEcb) Encrypt(plaintext []byte, key []byte) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = icrypto.ErrEncrypt.Errorf(err)
		}
	}()
	if ae.isHashKey {
		key = ae.hashKey(key)
	} else {
		err = icrypto.KeyLenCheck(key, ae.keyLens)
		if err != nil {
			return nil, err
		}
	}

	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	length := (len(plaintext) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, plaintext)
	pad := byte(len(plain) - len(plaintext))
	for i := len(plaintext); i < len(plain); i++ {
		plain[i] = pad
	}
	data = make([]byte, len(plain))
	for bs, be := 0, cipher.BlockSize(); bs <= len(plaintext); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(data[bs:be], plain[bs:be])
	}
	return data, nil
}

func (ae AesEcb) Decrypt(ciphertext []byte, key []byte) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = icrypto.ErrEncrypt.Errorf(err)
		}
	}()
	if ae.isHashKey {
		key = ae.hashKey(key)
	} else {
		err = icrypto.KeyLenCheck(key, ae.keyLens)
		if err != nil {
			return nil, err
		}
	}

	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	data = make([]byte, len(ciphertext))
	for bs, be := 0, cipher.BlockSize(); bs < len(ciphertext); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Decrypt(data[bs:be], ciphertext[bs:be])
	}
	trim := 0
	if len(data) > 0 {
		trim = len(data) - int(data[len(data)-1])
	}
	return data[:trim], nil
}

func (ae AesEcb) hashKey(key []byte) []byte {
	return mhash.ToHash(key)
}
