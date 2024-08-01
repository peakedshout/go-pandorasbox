package rsa

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/peakedshout/go-pandorasbox/pcrypto/icrypto"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
	"math/big"
)

var (
	PCryptoRsa           = Rsa{name: "rsa"}
	PCryptoRsaChunks     = Rsa{name: "rsa-chunks", chunks: true}
	PCryptoRsaCert       = RsaCert{name: "rsa-cert"}
	PCryptoRsaCertChunks = RsaCert{name: "rsa-cert-chunks", chunks: true}
)

type Rsa struct {
	name   string
	chunks bool
}

func (r Rsa) Name() string {
	return r.name
}

func (r Rsa) IsSymmetric() bool {
	return false
}

func (r Rsa) Encrypt(plaintext []byte, key []byte) (data []byte, err error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, icrypto.ErrPublicKeyInvalid.Errorf("not rsa public key")
	}

	if r.chunks {
		lens := len(plaintext)
		size := pub.Size() - 11
		bs := new(bytes.Buffer)
		for i := 0; i < len(plaintext); i += size {
			next := i + size
			if lens <= next {
				next = lens
			}
			b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, plaintext[i:next])
			if err != nil {
				return nil, err
			}
			bs.Write(b)
		}
		return bs.Bytes(), nil
	} else {
		return rsa.EncryptPKCS1v15(rand.Reader, pub, plaintext)
	}
}

func (r Rsa) Decrypt(ciphertext []byte, key []byte) (data []byte, err error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("private key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	if r.chunks {
		lens := len(ciphertext)
		size := priv.Size()
		bs := new(bytes.Buffer)
		for i := 0; i < len(ciphertext); i += size {
			next := i + size
			if lens <= next {
				next = lens
			}
			b, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext[i:next])
			if err != nil {
				return nil, err
			}
			bs.Write(b)
		}
		return bs.Bytes(), nil
	} else {
		return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	}
}

func (r Rsa) GenRsaKey(bits int) (public, private []byte, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	privateBuf := new(bytes.Buffer)
	err = pem.Encode(privateBuf, block)
	if err != nil {
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, nil, err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	publicBuf := new(bytes.Buffer)
	err = pem.Encode(publicBuf, block)
	if err != nil {
		return nil, nil, err
	}
	return publicBuf.Bytes(), privateBuf.Bytes(), nil
}

type RsaCert struct {
	name   string
	chunks bool
}

func (r RsaCert) Name() string {
	return r.name
}

func (r RsaCert) IsSymmetric() bool {
	return false
}

func (r RsaCert) Encrypt(plaintext []byte, key []byte) (data []byte, err error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("public key error")
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub, ok := c.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, icrypto.ErrPublicKeyInvalid.Errorf("not rsa public key")
	}

	if r.chunks {
		lens := len(plaintext)
		size := pub.Size() - 11
		bs := new(bytes.Buffer)
		for i := 0; i < len(plaintext); i += size {
			next := i + size
			if lens <= next {
				next = lens
			}
			b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, plaintext[i:next])
			if err != nil {
				return nil, err
			}
			bs.Write(b)
		}
		return bs.Bytes(), nil
	} else {
		return rsa.EncryptPKCS1v15(rand.Reader, pub, plaintext)
	}
}

func (r RsaCert) Decrypt(ciphertext []byte, key []byte) (data []byte, err error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("private key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	if r.chunks {
		lens := len(ciphertext)
		size := priv.Size()
		bs := new(bytes.Buffer)
		for i := 0; i < len(ciphertext); i += size {
			next := i + size
			if lens <= next {
				next = lens
			}
			b, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext[i:next])
			if err != nil {
				return nil, err
			}
			bs.Write(b)
		}
		return bs.Bytes(), nil
	} else {
		return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	}
}

func (r RsaCert) Sign(plaintext []byte, key []byte) (hashText []byte, sign []byte, err error) {
	h := mhash.HashSha512(plaintext)
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, nil, errors.New("public key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	sign, err = rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA512, h)
	if err != nil {
		return nil, nil, err
	}
	return h, sign, nil
}

func (r RsaCert) Verify(hashText []byte, sign []byte, key []byte) (ok bool, err error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return false, errors.New("public key error")
	}
	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, err
	}
	rsapub, ok := pub.PublicKey.(*rsa.PublicKey)
	if !ok {
		return false, icrypto.ErrPublicKeyInvalid.Errorf("not ecdsa public key")
	}
	err = rsa.VerifyPKCS1v15(rsapub, crypto.SHA512, hashText, sign)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r RsaCert) GenRsaCert(bits int, template *x509.Certificate) (cert, key []byte, err error) {
	if template == nil {
		max := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, max)
		if err != nil {
			return nil, nil, err
		}
		template = &x509.Certificate{
			SerialNumber: serialNumber,
		}
	}
	pk, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &pk.PublicKey, pk)
	if err != nil {
		return nil, nil, err
	}
	certBs := new(bytes.Buffer)
	err = pem.Encode(certBs, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, nil, err
	}
	keyBs := new(bytes.Buffer)
	err = pem.Encode(keyBs, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)})
	if err != nil {
		return nil, nil, err
	}
	return certBs.Bytes(), keyBs.Bytes(), nil
}
