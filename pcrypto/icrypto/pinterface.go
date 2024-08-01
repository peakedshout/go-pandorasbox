package icrypto

import (
	"sync"
)

type Interface interface {
	Name() string
	IsSymmetric() bool
	Encrypt(plaintext []byte, key []byte) (data []byte, err error)
	Decrypt(ciphertext []byte, key []byte) (data []byte, err error)
}

type DSAInterface interface {
	Name() string
	Sign(plaintext []byte, key []byte) (hashText []byte, sign []byte, err error)
	Verify(hashText []byte, sign []byte, key []byte) (ok bool, err error)
}

var gLock sync.Mutex
var gMap = make(map[string]Interface)
var gChange = false
var gList []Interface

func Register(pc Interface) {
	gLock.Lock()
	defer gLock.Unlock()
	gChange = true
	gMap[pc.Name()] = pc
}

// GetInterface This is only going to get package import and Register
func GetInterface(pt string) (Interface, error) {
	gLock.Lock()
	defer gLock.Unlock()
	pc, ok := gMap[pt]
	if !ok {
		return nil, ErrNotSupportType.Errorf(pt)
	}
	return pc, nil
}

// GetAllInterface This is only going to get package import and Register
func GetAllInterface() []Interface {
	gLock.Lock()
	defer gLock.Unlock()
	if gChange {
		list := make([]Interface, 0, len(gMap))
		for _, i := range gMap {
			list = append(list, i)
		}
		gList = list
		gChange = false
	}
	return gList
}

var AesKeyLens = []int{16, 24, 32}

func KeyLenCheck(key []byte, l []int) error {
	kl := len(key)
	for _, one := range l {
		if kl == one {
			return nil
		}
	}
	return ErrCheckKeyLen.Errorf(kl, l)
}
