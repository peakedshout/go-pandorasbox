package ecdsa

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/peakedshout/go-pandorasbox/pcrypto/icrypto"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
	"math/big"
)

var (
	PCryptoEcdsa     = Ecdsa{name: "ecdsa"}
	PCryptoEcdsaCert = EcdsaCert{name: "ecdsa-cert"}
)

type Ecdsa struct {
	name string
}

func (e Ecdsa) Name() string {
	return e.name
}

func (e Ecdsa) Sign(plaintext []byte, key []byte) (hashText []byte, sign []byte, err error) {
	h := mhash.HashSha512(plaintext)
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, nil, errors.New("private key error")
	}
	pirv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	sign, err = ecdsa.SignASN1(rand.Reader, pirv, h)
	if err != nil {
		return nil, nil, err
	}
	return h, sign, nil
}

func (e Ecdsa) Verify(hashText []byte, sign []byte, key []byte) (ok bool, err error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return false, errors.New("public key error")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, err
	}
	ecpub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return false, icrypto.ErrPublicKeyInvalid.Errorf("not ecdsa public key")
	}
	return ecdsa.VerifyASN1(ecpub, hashText, sign), nil
}

func (e Ecdsa) GenEcdsaKey(c elliptic.Curve) (public, private []byte, err error) {
	privateKey, err := ecdsa.GenerateKey(c, rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	derStream, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}
	block := &pem.Block{
		Type:  "ECDSA PRIVATE KEY",
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

type EcdsaCert struct {
	name string
}

func (e EcdsaCert) Name() string {
	return e.name
}

func (e EcdsaCert) Sign(plaintext []byte, key []byte) (hashText []byte, sign []byte, err error) {
	h := mhash.HashSha512(plaintext)
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, nil, errors.New("private key error")
	}
	pirv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	sign, err = ecdsa.SignASN1(rand.Reader, pirv, h)
	if err != nil {
		return nil, nil, err
	}
	return h, sign, nil
}

func (e EcdsaCert) Verify(hashText []byte, sign []byte, key []byte) (ok bool, err error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return false, errors.New("public key error")
	}
	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, err
	}
	ecpub, ok := pub.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return false, icrypto.ErrPublicKeyInvalid.Errorf("not ecdsa public key")
	}
	return ecdsa.VerifyASN1(ecpub, hashText, sign), nil
}

func (e EcdsaCert) GenEcdsaCert(c elliptic.Curve, template *x509.Certificate) (cert, key []byte, err error) {
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

	pk, err := ecdsa.GenerateKey(c, rand.Reader)
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
	pri, err := x509.MarshalECPrivateKey(pk)
	if err != nil {
		return nil, nil, err
	}
	keyBs := new(bytes.Buffer)
	err = pem.Encode(keyBs, &pem.Block{Type: "EC PRIVATE KEY", Bytes: pri})
	if err != nil {
		return nil, nil, err
	}
	return certBs.Bytes(), keyBs.Bytes(), nil
}
