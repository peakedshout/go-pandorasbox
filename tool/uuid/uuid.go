package uuid

import (
	"encoding/base64"
	"github.com/gofrs/uuid"
	"github.com/peakedshout/go-pandorasbox/tool/xbit"
	"math"
	"strings"
)

// NewId lens is 32 * n
func NewId(n int) (str string) {
	for i := n; i > 0; i-- {
		uid, err := uuid.NewV4()
		if err != nil {
			panic(err)
		}
		uidStr := uid.String()
		s := strings.Replace(uidStr, "-", "", -1)
		str += s
	}
	return
}

// NewIdn lens is n
func NewIdn(n int) (str string) {
	i := int(math.Ceil(float64(n) / 32))
	return NewId(i)[:n]
}

// NewBytes 16lens
func NewBytes() []byte {
	uid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return uid.Bytes()
}

func NewMMInt64() (max int64, min int64) {
	bytes := NewBytes()
	max, _ = xbit.BigFromBytes[int64](bytes[:8])
	min, _ = xbit.BigFromBytes[int64](bytes[8:16])
	return max, min
}

func NewStdBase64() string {
	return base64.RawStdEncoding.EncodeToString(NewBytes())
}

func NewUrlBase64() string {
	return base64.RawURLEncoding.EncodeToString(NewBytes())
}
