package aescrt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"github.com/peakedshout/go-pandorasbox/pcrypto/icrypto"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
)

var (
	PCryptoAesHashCrt = AesCrt{name: "aes-hash-crt", isHashKey: true}
	PCryptoAesXxxCrt  = AesCrt{name: "aes-xxx-crt", keyLens: icrypto.AesKeyLens}
	PCryptoAes128Crt  = AesCrt{name: "aes-128-crt", keyLens: []int{16}}
	PCryptoAes192Crt  = AesCrt{name: "aes-192-crt", keyLens: []int{24}}
	PCryptoAes256Crt  = AesCrt{name: "aes-256-crt", keyLens: []int{32}}
)

type AesCrt struct {
	name      string
	keyLens   []int
	isHashKey bool
}

func (ac AesCrt) Name() string {
	return ac.name
}

func (ac AesCrt) IsSymmetric() bool {
	return true
}

func (ac AesCrt) Encrypt(plaintext []byte, key []byte) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = icrypto.ErrEncrypt.Errorf(err)
		}
	}()
	if ac.isHashKey {
		key = ac.hashKey(key)
	} else {
		err = icrypto.KeyLenCheck(key, ac.keyLens)
		if err != nil {
			return nil, err
		}
	}

	return ac.handle(plaintext, key)
}

func (ac AesCrt) Decrypt(ciphertext []byte, key []byte) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = icrypto.ErrEncrypt.Errorf(err)
		}
	}()
	if ac.isHashKey {
		key = ac.hashKey(key)
	} else {
		err = icrypto.KeyLenCheck(key, ac.keyLens)
		if err != nil {
			return nil, err
		}
	}

	return ac.handle(ciphertext, key)
}

func (ac AesCrt) handle(plaintext []byte, key []byte) (data []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := bytes.Repeat([]byte("1"), block.BlockSize())
	stream := cipher.NewCTR(block, iv)
	data = make([]byte, len(plaintext))
	stream.XORKeyStream(data, plaintext)
	return data, nil
}

func (ac AesCrt) hashKey(key []byte) []byte {
	return mhash.ToHash(key)
}
