package message

import "encoding/binary"

const (
	HEADER_LEN = 11
)

type Header struct {
	Type      uint16
	Timestamp uint64
	Encrypted bool
}

func (self *Header) serialize() []byte {
	res := make([]byte, HEADER_LEN)
	binary.BigEndian.PutUint16(res, self.Type)
	binary.BigEndian.PutUint64(res[2:], self.Timestamp)
	if self.Encrypted {
		res[10] = 1
	}
	return res
}

func deserializeHeader(header []byte) Header {
	typ := binary.BigEndian.Uint16(header[:2])
	timestamp := binary.BigEndian.Uint64(header[2:10])
	encrypted := header[10] == 1
	return Header{typ, timestamp, encrypted}
}
