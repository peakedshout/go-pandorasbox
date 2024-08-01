package pcrypto

import (
	"bytes"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aesgcm"
	"github.com/peakedshout/go-pandorasbox/pcrypto/rsa"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"testing"
)

func TestCryptoEmpty(t *testing.T) {
	data := []byte(uuid.NewId(1))
	pc, err := NewPCrypto(nil)
	if err != nil {
		t.Error(err)
	}
	b, err := pc.Encrypt(data)
	if err != nil {
		t.Error(err)
	}
	b, err = pc.Decrypt(b)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(data, b) {
		t.Error()
	}
}

func TestCryptoSymmetric(t *testing.T) {
	data := []byte(uuid.NewId(1))
	key := []byte(uuid.NewId(1))
	pc, err := NewPCrypto(aesgcm.PCryptoAes256Gcm, key)
	if err != nil {
		t.Error(err)
	}
	b, err := pc.Encrypt(data)
	if err != nil {
		t.Error(err)
	}
	b, err = pc.Decrypt(b)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(data, b) {
		t.Error()
	}
}

func TestCryptoAsymmetric(t *testing.T) {
	data := []byte(uuid.NewId(1))
	pub, pri, err := rsa.PCryptoRsa.GenRsaKey(1024)
	if err != nil {
		t.Error(err)
	}
	pc, err := NewPCrypto(rsa.PCryptoRsa, pub, pri)
	if err != nil {
		t.Error(err)
	}
	b, err := pc.Encrypt(data)
	if err != nil {
		t.Error(err)
	}
	b, err = pc.Decrypt(b)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(data, b) {
		t.Error()
	}
}
