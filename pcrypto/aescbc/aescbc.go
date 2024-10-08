package aescbc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"github.com/peakedshout/go-pandorasbox/pcrypto/icrypto"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
)

var (
	PCryptoAesHashCbc = AesCbc{name: "aes-hash-cbc", isHashKey: true}
	PCryptoAesXxxCbc  = AesCbc{name: "aes-xxx-cbc", keyLens: icrypto.AesKeyLens}
	PCryptoAes128Cbc  = AesCbc{name: "aes-128-cbc", keyLens: []int{16}}
	PCryptoAes192Cbc  = AesCbc{name: "aes-192-cbc", keyLens: []int{24}}
	PCryptoAes256Cbc  = AesCbc{name: "aes-256-cbc", keyLens: []int{32}}
)

type AesCbc struct {
	name      string
	keyLens   []int
	isHashKey bool
}

func (ac AesCbc) Name() string {
	return ac.name
}

func (ac AesCbc) IsSymmetric() bool {
	return true
}

func (ac AesCbc) Encrypt(plaintext []byte, key []byte) (data []byte, err error) {
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

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	plaintext = ac.toPKCS7Padding(plaintext, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	data = make([]byte, len(plaintext))
	blockMode.CryptBlocks(data, plaintext)
	return data, nil
}

func (ac AesCbc) Decrypt(ciphertext []byte, key []byte) (data []byte, err error) {
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

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	orig := make([]byte, len(ciphertext))
	blockMode.CryptBlocks(orig, ciphertext)
	data = ac.toPKCS7UnPadding(orig)
	return data, nil
}

func (ac AesCbc) toPKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (ac AesCbc) toPKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func (ac AesCbc) hashKey(key []byte) []byte {
	return mhash.ToHash(key)
}
