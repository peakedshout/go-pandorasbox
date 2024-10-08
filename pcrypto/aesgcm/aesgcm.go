package aesgcm

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
	PCryptoAesHashGcm = AesGcm{name: "aes-hash-gcm", isHashKey: true}
	PCryptoAesXxxGcm  = AesGcm{name: "aes-xxx-gcm", keyLens: icrypto.AesKeyLens}
	PCryptoAes128Gcm  = AesGcm{name: "aes-128-gcm", keyLens: []int{16}}
	PCryptoAes192Gcm  = AesGcm{name: "aes-192-gcm", keyLens: []int{24}}
	PCryptoAes256Gcm  = AesGcm{name: "aes-256-gcm", keyLens: []int{32}}
)

type AesGcm struct {
	name      string
	keyLens   []int
	isHashKey bool
}

func (ag AesGcm) Name() string {
	return ag.name
}

func (ag AesGcm) IsSymmetric() bool {
	return true
}

func (ag AesGcm) Encrypt(plaintext []byte, key []byte) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = icrypto.ErrEncrypt.Errorf(err)
		}
	}()
	if ag.isHashKey {
		key = ag.hashKey(key)
	} else {
		err = icrypto.KeyLenCheck(key, ag.keyLens)
		if err != nil {
			return nil, err
		}
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (ag AesGcm) Decrypt(ciphertext []byte, key []byte) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = icrypto.ErrDecrypt.Errorf(err)
		}
	}()
	if ag.isHashKey {
		key = ag.hashKey(key)
	} else {
		err = icrypto.KeyLenCheck(key, ag.keyLens)
		if err != nil {
			return nil, err
		}
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (ag AesGcm) hashKey(key []byte) []byte {
	return mhash.ToHash(key)
}
