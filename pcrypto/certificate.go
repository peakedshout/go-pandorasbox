package pcrypto

import (
	"crypto/tls"
	"github.com/peakedshout/go-pandorasbox/pcrypto/rsa"
)

func MakeTlsConfigFromFile(certFile string, keyFile string) (*tls.Config, error) {
	c, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{c}, InsecureSkipVerify: true}
	return cfg, nil
}

func MustMakeTlsConfig(cert []byte, key []byte) *tls.Config {
	config, err := MakeTlsConfig(cert, key)
	if err != nil {
		panic(err)
	}
	return config
}

func MakeTlsConfig(cert []byte, key []byte) (*tls.Config, error) {
	c, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{c}, InsecureSkipVerify: true}
	return cfg, nil
}

func NewDefaultTlsConfigWithRaw() (*tls.Config, []byte, []byte, error) {
	cert, key, err := rsa.PCryptoRsaCert.GenRsaCert(1024, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	c, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, nil, nil, err
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{c}, InsecureSkipVerify: true}
	return cfg, cert, key, nil
}

func NewDefaultTlsConfig() (*tls.Config, error) {
	cert, key, err := rsa.PCryptoRsaCert.GenRsaCert(1024, nil)
	if err != nil {
		return nil, err
	}
	c, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{c}, InsecureSkipVerify: true}
	return cfg, nil
}

func MustNewDefaultTlsConfig() *tls.Config {
	config, err := NewDefaultTlsConfig()
	if err != nil {
		panic(err)
	}
	return config
}
