package pcrypto

import (
	"bytes"
	"github.com/peakedshout/go-pandorasbox/pcrypto/icrypto"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
)

type PCrypto interface {
	Encrypt(plaintext []byte) (data []byte, err error)
	Decrypt(ciphertext []byte) (data []byte, err error)
	Name() string
	Hash() []byte
	IsSymmetric() bool
}

var CryptoPlaintext = &CryptoEmpty{}

// NewPCrypto
//
//	Symmetric len(keys) must 1; keys[0] will be SymmetricKey
//	Asymmetric len(keys) must 2; keys[0] will be PublicKey, keys[1] will be PrivateKey
func NewPCrypto(c icrypto.Interface, keys ...[]byte) (PCrypto, error) {
	if c == nil {
		ce := &CryptoEmpty{}
		return ce, nil
	}
	l := len(keys)
	if c.IsSymmetric() {
		if l != 1 {
			return nil, icrypto.ErrUnexpectedKeys.Errorf(l, 1)
		}
		cs := &CryptoSymmetric{
			symmetricKey: keys[0],
			hash:         mhash.ToHash(bytes.Join([][]byte{keys[0], keys[0]}, nil)),
			crypto:       c,
		}
		return cs, nil
	} else {
		if l != 2 {
			return nil, icrypto.ErrUnexpectedKeys.Errorf(l, 2)
		}
		ca := &CryptoAsymmetric{
			publicKey:  keys[0],
			privateKey: keys[1],
			hash:       mhash.ToHash(bytes.Join([][]byte{keys[0], keys[0]}, nil)),
			crypto:     c,
		}
		return ca, nil
	}
}

func MustNewPCrypto(c icrypto.Interface, keys ...[]byte) PCrypto {
	crypto, err := NewPCrypto(c, keys...)
	if err != nil {
		panic(err)
	}
	return crypto
}

type CryptoEmpty struct {
}

func (pe *CryptoEmpty) Encrypt(plaintext []byte) (data []byte, err error) {
	return plaintext, nil
}

func (pe *CryptoEmpty) Decrypt(ciphertext []byte) (data []byte, err error) {
	return ciphertext, nil
}

func (pe *CryptoEmpty) Name() string {
	return "_"
}

func (pe *CryptoEmpty) Hash() []byte {
	return nil
}

func (pe *CryptoEmpty) IsSymmetric() bool {
	return true
}

type CryptoSymmetric struct {
	symmetricKey []byte
	hash         []byte
	crypto       icrypto.Interface
}

func (pc *CryptoSymmetric) Encrypt(plaintext []byte) (data []byte, err error) {
	return pc.crypto.Encrypt(plaintext, pc.symmetricKey)
}

func (pc *CryptoSymmetric) Decrypt(ciphertext []byte) (data []byte, err error) {
	return pc.crypto.Decrypt(ciphertext, pc.symmetricKey)
}

func (pc *CryptoSymmetric) Name() string {
	return pc.crypto.Name()
}

func (pc *CryptoSymmetric) Hash() []byte {
	return pc.hash
}

func (pc *CryptoSymmetric) IsSymmetric() bool {
	return pc.crypto.IsSymmetric()
}

type CryptoAsymmetric struct {
	publicKey  []byte
	privateKey []byte
	hash       []byte
	crypto     icrypto.Interface
}

func (pa *CryptoAsymmetric) Encrypt(plaintext []byte) (data []byte, err error) {
	return pa.crypto.Encrypt(plaintext, pa.publicKey)
}

func (pa *CryptoAsymmetric) Decrypt(ciphertext []byte) (data []byte, err error) {
	return pa.crypto.Decrypt(ciphertext, pa.privateKey)
}

func (pa *CryptoAsymmetric) Name() string {
	return pa.crypto.Name()
}

func (pa *CryptoAsymmetric) Hash() []byte {
	return pa.hash
}

func (pa *CryptoAsymmetric) IsSymmetric() bool {
	return pa.crypto.IsSymmetric()
}
