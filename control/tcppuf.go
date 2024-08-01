package control

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/peakedshout/go-pandorasbox/tool/xbit"
	"net"
	"time"
)

type tcpMessage struct {
	*TCPHeader
	data []byte
}

func unmarshalTcpMessage(b []byte) (*tcpMessage, error) {
	header, err := ParseTCPHeader(b)
	if err != nil {
		return nil, err
	}
	data := b[header.HeaderSize():]
	return &tcpMessage{
		TCPHeader: header,
		data:      data,
	}, nil
}

func newTcpMessage(src, dst uint16, data []byte) *tcpMessage {
	return &tcpMessage{
		TCPHeader: &TCPHeader{
			Source:      src,
			Destination: dst,
		},
		data: data,
	}
}

func (tm *tcpMessage) Bytes() ([]byte, error) {
	th, err := tm.TCPHeader.Marshal()
	if err != nil {
		return nil, err
	}
	var bs bytes.Buffer
	buf := bytes.Join([][]byte{th, tm.data}, nil)
	bs.Write(buf[:16])
	err = binary.Write(&bs, binary.BigEndian, Csum(buf, tm.SrcIp, tm.DstIp))
	if err != nil {
		return nil, err
	}
	bs.Write(buf[18:])
	return bs.Bytes(), nil
}

var (
	errTcpHeaderTooShort = errors.New("tcp header too short")
	errTcpHeaderTooLong  = errors.New("tcp header too long")
)

const (
	FIN = 1  // 00 0001
	SYN = 2  // 00 0010
	RST = 4  // 00 0100
	PSH = 8  // 00 1000
	ACK = 16 // 01 0000
	URG = 32 // 10 0000
)

const baseTcpHeaderSize = 20
const maxTcpHeaderSize = 60

type TCPHeader struct {
	SrcIp net.IP
	DstIp net.IP

	Source      uint16
	Destination uint16
	SeqNum      uint32
	AckNum      uint32
	DataOffset  uint8 // 4 bits
	Reserved    uint8 // 3 bits
	ECN         uint8 // 3 bits
	Ctrl        uint8 // 6 bits
	Window      uint16
	Checksum    uint16 // Kernel will set this if it's 0
	Urgent      uint16
	Options     []TCPOption
	RawOptions  []byte
}

type TCPOption struct {
	Kind   uint8
	Length uint8
	Data   []byte
}

func ParseTCPHeader(data []byte) (*TCPHeader, error) {
	var tcp TCPHeader
	r := bytes.NewReader(data)
	err := binary.Read(r, binary.BigEndian, &tcp.Source)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.BigEndian, &tcp.Destination)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.BigEndian, &tcp.SeqNum)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.BigEndian, &tcp.AckNum)
	if err != nil {
		return nil, err
	}

	var mix uint16
	err = binary.Read(r, binary.BigEndian, &mix)
	if err != nil {
		return nil, err
	}
	tcp.DataOffset = byte(mix >> 12)  // top 4 bits
	tcp.Reserved = byte(mix >> 9 & 7) // 3 bits
	tcp.ECN = byte(mix >> 6 & 7)      // 3 bits
	tcp.Ctrl = byte(mix & 0x3f)       // bottom 6 bits

	if int(tcp.DataOffset) > len(data) {
		return nil, errTcpHeaderTooShort
	}

	err = binary.Read(r, binary.BigEndian, &tcp.Window)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.BigEndian, &tcp.Checksum)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.BigEndian, &tcp.Urgent)
	if err != nil {
		return nil, err
	}

	tcp.RawOptions = data[baseTcpHeaderSize:tcp.HeaderSize()]

	return &tcp, nil
}

func (tcp *TCPHeader) HasFlag(flagBit byte) bool {
	return tcp.Ctrl&flagBit != 0
}

func (tcp *TCPHeader) SetFlag(flagBit byte, flag bool) {
	switch flagBit {
	case FIN:
		tcp.Ctrl = xbit.SetFlag[uint8](tcp.Ctrl, 1, flag)
	case SYN:
		tcp.Ctrl = xbit.SetFlag[uint8](tcp.Ctrl, 2, flag)
	case RST:
		tcp.Ctrl = xbit.SetFlag[uint8](tcp.Ctrl, 3, flag)
	case PSH:
		tcp.Ctrl = xbit.SetFlag[uint8](tcp.Ctrl, 4, flag)
	case ACK:
		tcp.Ctrl = xbit.SetFlag[uint8](tcp.Ctrl, 5, flag)
	case URG:
		tcp.Ctrl = xbit.SetFlag[uint8](tcp.Ctrl, 6, flag)
	}
}

func (tcp *TCPHeader) HeaderSize() int {
	return int(tcp.DataOffset) * 4 //offset * 32 / 8 to bytes
}

//func (tcp *TCPHeader) String() string {
//	if tcp == nil {
//		return "<nil>"
//	}
//	return fmt.Sprintf("Source=%v Destination=%v SeqNum=%v AckNum=%v DataOffset=%v Reserved=%v ECN=%v Ctrl=%v Window=%v Checksum=%v Urgent=%v Options=%v", tcp.Source, tcp.Destination, tcp.SeqNum, tcp.AckNum, tcp.DataOffset, tcp.Reserved, tcp.ECN, tcp.Ctrl, tcp.Window, tcp.Checksum, tcp.Options)
//}

func (tcp *TCPHeader) Marshal() ([]byte, error) {
	optBuf := new(bytes.Buffer)
	for _, option := range tcp.Options {
		err := binary.Write(optBuf, binary.BigEndian, option.Kind)
		if err != nil {
			return nil, err
		}
		if option.Length > 1 {
			err = binary.Write(optBuf, binary.BigEndian, option.Length)
			if err != nil {
				return nil, err
			}
			err = binary.Write(optBuf, binary.BigEndian, option.Data)
			if err != nil {
				return nil, err
			}
		}
	}
	if len(tcp.RawOptions) != 0 {
		optBuf.Write(tcp.RawOptions)
	}
	for optBuf.Len()%4 != 0 {
		optBuf.WriteByte(0)
	}

	if optBuf.Len() > maxTcpHeaderSize-baseTcpHeaderSize {
		return nil, errTcpHeaderTooLong
	}
	tcp.DataOffset = uint8((baseTcpHeaderSize + optBuf.Len()) / 4)

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, tcp.Source)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, tcp.Destination)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, tcp.SeqNum)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, tcp.AckNum)
	if err != nil {
		return nil, err
	}

	var mix uint16
	mix = uint16(tcp.DataOffset)<<12 | // top 4 bits
		uint16(tcp.Reserved)<<9 | // 3 bits
		uint16(tcp.ECN)<<6 | // 3 bits
		uint16(tcp.Ctrl) // bottom 6 bits
	err = binary.Write(buf, binary.BigEndian, mix)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, tcp.Window)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, tcp.Checksum)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, tcp.Urgent)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, optBuf.Bytes())
	if err != nil {
		return nil, err
	}

	out := buf.Bytes()

	// Pad to min tcp header size, which is 20 bytes (5 32-bit words)
	pad := baseTcpHeaderSize - len(out)
	for i := 0; i < pad; i++ {
		out = append(out, 0)
	}

	return out, nil
}

// Csum TCP Checksum
func Csum(data []byte, srcip, dstip []byte) uint16 {
	var pseudoHeader bytes.Buffer
	pseudoHeader.Write(srcip)
	pseudoHeader.Write(dstip)
	pseudoHeader.Write([]byte{
		0,                  // zero
		6,                  // protocol number (6 == TCP)
		0, byte(len(data)), // TCP length (16 bits), not inc pseudo header
	})

	sumThis := make([]byte, 0, pseudoHeader.Len()+len(data))
	sumThis = append(sumThis, pseudoHeader.Bytes()...)
	sumThis = append(sumThis, data...)
	//fmt.Printf("% x\n", sumThis)

	lenSumThis := len(sumThis)
	var nextWord uint16
	var sum uint32
	for i := 0; i+1 < lenSumThis; i += 2 {
		nextWord = uint16(sumThis[i])<<8 | uint16(sumThis[i+1])
		sum += uint32(nextWord)
	}
	if lenSumThis%2 != 0 {
		//fmt.Println("Odd byte")
		sum += uint32(sumThis[len(sumThis)-1])
	}

	// Add back any carry, and any carry from adding the carry
	sum = (sum >> 16) + (sum & 0xffff)
	sum = sum + (sum >> 16)

	// Bitwise complement
	return uint16(^sum)
}

func DefaultRawOptions() []byte {
	var bs bytes.Buffer
	bs.Write([]byte{
		0x2, 0x4, 0x5, 0xb4, // maximun-segment-size option
		0x1, 0x3, 0x3, 0x6, //	window scale factor option
		0x1, 0x1, 0x8, 0xa, //	timestamp option
	})
	_ = binary.Write(&bs, binary.BigEndian, uint32(time.Now().Unix())) // timestamp send
	bs.Write([]byte{
		0x0, 0x0, 0x0, 0x0, // timestamp ack zero
		0x4, 0x2, 0x0, 0x0, // SACK-permitted option
	})
	return bs.Bytes()
}
