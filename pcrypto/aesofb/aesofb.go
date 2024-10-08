package aesofb

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/pcrypto/icrypto"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
	"io"
)

var (
	PCryptoAesHashOfb = AesOfb{name: "aes-hash-ofb", isHashKey: true}
	PCryptoAesXxxOfb  = AesOfb{name: "aes-xxx-ofb", keyLens: icrypto.AesKeyLens}
	PCryptoAes128Ofb  = AesOfb{name: "aes-128-ofb", keyLens: []int{16}}
	PCryptoAes192Ofb  = AesOfb{name: "aes-192-ofb", keyLens: []int{24}}
	PCryptoAes256Ofb  = AesOfb{name: "aes-256-ofb", keyLens: []int{32}}
)

type AesOfb struct {
	name      string
	keyLens   []int
	isHashKey bool
}

func (ao AesOfb) Name() string {
	return ao.name
}

func (ao AesOfb) IsSymmetric() bool {
	return true
}

func (ao AesOfb) Encrypt(plaintext []byte, key []byte) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = icrypto.ErrEncrypt.Errorf(err)
		}
	}()
	if ao.isHashKey {
		key = ao.hashKey(key)
	} else {
		err = icrypto.KeyLenCheck(key, ao.keyLens)
		if err != nil {
			return nil, err
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext = ao.toPKCS7Padding(plaintext, aes.BlockSize)
	data = make([]byte, aes.BlockSize+len(plaintext))
	iv := data[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(data[aes.BlockSize:], plaintext)
	return data, nil
}

func (ao AesOfb) Decrypt(ciphertext []byte, key []byte) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = icrypto.ErrEncrypt.Errorf(err)
		}
	}()
	if ao.isHashKey {
		key = ao.hashKey(key)
	} else {
		err = icrypto.KeyLenCheck(key, ao.keyLens)
		if err != nil {
			return nil, err
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("data is not a multiple of the block size")
	}
	data = make([]byte, len(ciphertext))
	mode := cipher.NewOFB(block, iv)
	mode.XORKeyStream(data, ciphertext)
	data = ao.toPKCS7UnPadding(data)
	return data, nil
}

func (ao AesOfb) toPKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (ao AesOfb) toPKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func (ao AesOfb) hashKey(key []byte) []byte {
	return mhash.ToHash(key)
}
