package rsa

import (
	"bytes"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"testing"
)

func Test(t *testing.T) {
	pub, pri, err := PCryptoRsa.GenRsaKey(1024)
	if err != nil {
		t.Error(err)
	}
	data := []byte(uuid.NewIdn(117))
	b, err := PCryptoRsa.Encrypt(data, pub)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoRsa.Decrypt(b, pri)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(data, b) {
		t.Failed()
	}

	data = []byte(uuid.NewIdn(1024))
	b, err = PCryptoRsaChunks.Encrypt(data, pub)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoRsaChunks.Decrypt(b, pri)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(data, b) {
		t.Failed()
	}
}

func TestCert(t *testing.T) {
	cert, key, err := PCryptoRsaCert.GenRsaCert(1024, nil)
	if err != nil {
		t.Error(err)
	}
	data := []byte(uuid.NewIdn(117))
	b, err := PCryptoRsaCert.Encrypt(data, cert)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoRsaCert.Decrypt(b, key)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(data, b) {
		t.Failed()
	}

	data = []byte(uuid.NewIdn(1024))
	b, err = PCryptoRsaCertChunks.Encrypt(data, cert)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoRsaCertChunks.Decrypt(b, key)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(data, b) {
		t.Failed()
	}
}

func TestGenRsaCert(t *testing.T) {
	cert, key, err := PCryptoRsaCert.GenRsaCert(4096, nil)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(cert))
	fmt.Println(string(key))
}

func TestGenRsaKey(t *testing.T) {
	pub, pri, err := PCryptoRsa.GenRsaKey(4096)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(pub))
	fmt.Println(string(pri))
}
