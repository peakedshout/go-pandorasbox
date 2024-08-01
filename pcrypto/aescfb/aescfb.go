package aescfb

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"github.com/peakedshout/go-pandorasbox/pcrypto/icrypto"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
	"io"
)

var (
	PCryptoAesHashCfb = AesCfb{name: "aes-hash-cfb", isHashKey: true}
	PCryptoAesXxxCfb  = AesCfb{name: "aes-xxx-cfb", keyLens: icrypto.AesKeyLens}
	PCryptoAes128Cfb  = AesCfb{name: "aes-128-cfb", keyLens: []int{16}}
	PCryptoAes192Cfb  = AesCfb{name: "aes-192-cfb", keyLens: []int{24}}
	PCryptoAes256Cfb  = AesCfb{name: "aes-256-cfb", keyLens: []int{32}}
)

type AesCfb struct {
	name      string
	keyLens   []int
	isHashKey bool
}

func (ae AesCfb) Name() string {
	return ae.name
}

func (ae AesCfb) IsSymmetric() bool {
	return true
}

func (ae AesCfb) Encrypt(plaintext []byte, key []byte) (data []byte, err error) {
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

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	data = make([]byte, aes.BlockSize+len(plaintext))
	iv := data[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(data[aes.BlockSize:], plaintext)
	return data, nil
}

func (ae AesCfb) Decrypt(ciphertext []byte, key []byte) (data []byte, err error) {
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

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return data, nil
}

func (ae AesCfb) hashKey(key []byte) []byte {
	return mhash.ToHash(key)
}
