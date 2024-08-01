package ecdsa

import (
	"crypto/elliptic"
	"fmt"
	"testing"
)

func TestEcdsa(t *testing.T) {
	public, private, err := PCryptoEcdsa.GenEcdsaKey(elliptic.P256())
	if err != nil {
		t.Error(err)
	}
	data := []byte("ddwdwdwdw")
	text, s, err := PCryptoEcdsa.Sign(data, private)
	if err != nil {
		t.Error(err)
	}
	ok, err := PCryptoEcdsa.Verify(text, s, public)
	if !ok || err != nil {
		t.Error(err)
	}
}

func TestEcdsaCert(t *testing.T) {
	public, private, err := PCryptoEcdsaCert.GenEcdsaCert(elliptic.P256(), nil)
	if err != nil {
		t.Error(err)
	}
	data := []byte("ddwdwdwdw")
	text, s, err := PCryptoEcdsaCert.Sign(data, private)
	if err != nil {
		t.Error(err)
	}
	ok, err := PCryptoEcdsaCert.Verify(text, s, public)
	if !ok || err != nil {
		t.Error(err)
	}
}

func TestGenEcdsaKey(t *testing.T) {
	public, private, err := PCryptoEcdsa.GenEcdsaKey(elliptic.P256())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(public))
	fmt.Println(string(private))
}

func TestGenEcdsaCert(t *testing.T) {
	public, private, err := PCryptoEcdsaCert.GenEcdsaCert(elliptic.P256(), nil)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(public))
	fmt.Println(string(private))
}
