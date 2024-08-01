package cfcprotocol

import (
	"bytes"
	"encoding/binary"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
	"github.com/peakedshout/go-pandorasbox/uerror"
	"io"
)

var CFCPlaintext = NewCFCProtocol(pcrypto.CryptoPlaintext)

type CFCProtocol struct {
	crypto pcrypto.PCrypto
}

func NewCFCProtocol(c pcrypto.PCrypto) *CFCProtocol {
	return &CFCProtocol{crypto: c}
}

func (cp *CFCProtocol) Encode(writer io.Writer, a any) error {
	bs, err := cp.EncodeBytes(a)
	if err != nil {
		return err
	}
	_, err = writer.Write(bs)
	return err
}

func (cp *CFCProtocol) Decode(reader io.Reader, a any) error {
	bs := new(bytes.Buffer)
	for {
		b, err := cp.parseBuffers(reader)
		if err == nil || uerror.Is(err, errCFCProtocolWaitPacket) {
			bs.Write(b)
			if err == nil {
				break
			}
		} else {
			return err
		}
	}
	data, err := cp.crypto.Decrypt(bs.Bytes())
	if err != nil {
		return err
	}
	return decode(data, a)
}

func (cp *CFCProtocol) EncodeBytes(a any) ([]byte, error) {
	bk, err := encode(a)
	if err != nil {
		return nil, err
	}
	if cp.crypto != nil {
		bk, err = cp.crypto.Encrypt(bk)
		if err != nil {
			return nil, err
		}
	}
	size := BufferSize - getHeaderSize()

	lens := len(bk)
	bs := make([][]byte, 0, (lens/size)+1) // equal or +1

	for j := 0; j < lens; j += size {
		next := j + size
		if lens <= next {
			next = lens
		}
		bs = append(bs, bk[j:next])
	}
	return cp.assemblyBuffers(bs)
}

func (cp *CFCProtocol) DecodeBytes(b []byte, a any) error {
	r := bytes.NewReader(b)
	return cp.Decode(r, a)
}

func (cp *CFCProtocol) assemblyBuffers(bs [][]byte) ([]byte, error) {
	l := len(bs)
	bo := make([][]byte, 0, l)
	for i, b := range bs {
		lens := int64(len(b) + getHeaderSize()) //ver len hash num... data
		var pkg = new(bytes.Buffer)
		//version
		err := binary.Write(pkg, binary.LittleEndian, []byte(version))
		if err != nil {
			return nil, err
		}
		//len
		err = binary.Write(pkg, binary.LittleEndian, lens)
		if err != nil {
			return nil, err
		}
		//null
		err = binary.Write(pkg, binary.LittleEndian, getNull())
		if err != nil {
			return nil, err
		}
		//num+data
		var pkg2 = new(bytes.Buffer)
		err = binary.Write(pkg2, binary.LittleEndian, int64(l-i-1))
		if err != nil {
			return nil, err
		}
		err = binary.Write(pkg2, binary.LittleEndian, b)
		if err != nil {
			return nil, err
		}
		b2 := pkg2.Bytes()
		//hash
		h, err := cp.makeHash(b2)
		if err != nil {
			return nil, err
		}
		err = binary.Write(pkg, binary.LittleEndian, h)
		if err != nil {
			return nil, err
		}
		//num + data
		err = binary.Write(pkg, binary.LittleEndian, b2)
		if err != nil {
			panic(err)
		}
		bo = append(bo, pkg.Bytes())
	}
	return bytes.Join(bo, nil), nil
}

func (cp *CFCProtocol) parseBuffers(reader io.Reader) (b []byte, err error) {
	header := make([]byte, getHeaderSize())
	_, err = io.ReadFull(reader, header)
	if err != nil {
		return nil, err
	}

	// version
	ver := header[:versionSize]
	if string(ver) != version {
		err = ErrCFCProtocolIsNotGoCFC.Errorf(string(ver), version)
		return nil, err
	}
	// lens
	lenb := header[versionSize : versionSize+lenSize]
	lengBuff := bytes.NewBuffer(lenb)
	var lens int64
	err = binary.Read(lengBuff, binary.LittleEndian, &lens)
	if err != nil {
		return nil, err
	}
	if lens <= int64(getHeaderSize()) {
		err = ErrCFCProtocolLensTooShort.Errorf(lens, getHeaderSize())
		return nil, err
	}
	if lens > BufferSize {
		err = ErrCFCProtocolLensTooLong.Errorf(lens, BufferSize)
		return nil, err
	}
	data := make([]byte, lens-int64(getHeaderSize()))
	_, err = io.ReadFull(reader, data)
	if err != nil {
		return nil, err
	}
	// hash
	h := header[versionSize+lenSize+nullSize : versionSize+lenSize+nullSize+hashSize]
	h2, err := cp.makeHash(bytes.Join([][]byte{header[versionSize+lenSize+nullSize+hashSize:], data}, nil))
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(h, h2) {
		err = ErrCFCProtocolHashCheckFailed.Errorf()
		return nil, err
	}
	// num
	lengBuff2 := bytes.NewBuffer(header[versionSize+lenSize+nullSize+hashSize : versionSize+lenSize+nullSize+hashSize+numSize])
	var num int64
	err = binary.Read(lengBuff2, binary.LittleEndian, &num)
	if err != nil {
		return nil, err
	}
	if num != 0 {
		return data, errCFCProtocolWaitPacket.Errorf()
	}
	return data, nil
}

func (cp *CFCProtocol) makeHash(b []byte) (h []byte, err error) {
	if cp.crypto != nil {
		b = append(b, cp.crypto.Hash()...)
	}
	h = mhash.ToHash(b)
	return h, nil
}
