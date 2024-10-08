package cfcprotocol

const ProtocolName = `cfc`

const version = `go-CFC-v02.00.00`

// bytes size = 4096
// 4028 = 4096 - 16 - 8 - 4 - 32 - 8
const (
	BufferSize  = 4096 // packet max size
	versionSize = 16   // cfc protocol version size
	lenSize     = 8    // packet len size
	nullSize    = 4    // null size
	hashSize    = 32   // packet hash size
	numSize     = 8    //

	dataSize = BufferSize - versionSize - lenSize - nullSize - hashSize - numSize
)

func getHeaderSize() int {
	return versionSize + lenSize + nullSize + hashSize + numSize
}

func GetDataSize() int {
	return dataSize
}

func getNull() []byte {
	return make([]byte, 4)
}
