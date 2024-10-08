package mhash

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
)

// ToHash sha256 hash
func ToHash(b []byte) []byte {
	s := sha256.New()
	s.Write(b)
	return s.Sum(nil)
}

// CheckHash sha256 hash
func CheckHash(h []byte, data []byte) bool {
	h2 := ToHash(data)
	if len(h) != len(h2) {
		return false
	}
	return bytes.Equal(h, h2)
}

func HashMd5(b []byte, sums ...[]byte) []byte {
	s := md5.New()
	s.Write(b)
	return s.Sum(bytes.Join(sums, nil))
}

func HashSha1(b []byte, sums ...[]byte) []byte {
	s := sha1.New()
	s.Write(b)
	return s.Sum(bytes.Join(sums, nil))
}

func HashSha256(b []byte, sums ...[]byte) []byte {
	s := sha256.New()
	s.Write(b)
	return s.Sum(bytes.Join(sums, nil))
}

func HashSha512_224(b []byte, sums ...[]byte) []byte {
	s := sha512.New512_224()
	s.Write(b)
	return s.Sum(bytes.Join(sums, nil))
}

func HashSha512_256(b []byte, sums ...[]byte) []byte {
	s := sha512.New512_256()
	s.Write(b)
	return s.Sum(bytes.Join(sums, nil))
}

func HashSha512_384(b []byte, sums ...[]byte) []byte {
	s := sha512.New384()
	s.Write(b)
	return s.Sum(bytes.Join(sums, nil))
}

func HashSha512(b []byte, sums ...[]byte) []byte {
	s := sha512.New()
	s.Write(b)
	return s.Sum(bytes.Join(sums, nil))
}
