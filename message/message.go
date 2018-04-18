package message

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/golemfactory/bootstrap_go/cbor"
)

const (
	SIG_LEN = 65
)

type Message interface {
	GetType() uint16

	shouldEncrypt() bool

	getSignature() []byte
	setSignature(sig []byte)

	getTimestamp() uint64
	setTimestamp(ts uint64)

	serializationExtraData() []byte
}

type baseMessage struct {
	Timestamp uint64
	Sig       []byte
}

func (self *baseMessage) getSignature() []byte {
	return self.Sig
}

func (self *baseMessage) setSignature(sig []byte) {
	self.Sig = sig
}

func (self *baseMessage) getTimestamp() uint64 {
	return self.Timestamp
}

func (self *baseMessage) setTimestamp(ts uint64) {
	self.Timestamp = ts
}

func (self *baseMessage) serializationExtraData() []byte {
	return []byte{}
}

type EncryptFunc = func([]byte) ([]byte, error)
type SignFunc = func([]byte) ([]byte, error)

func Serialize(msg Message, encrypt EncryptFunc, sign SignFunc) ([]byte, error) {
	header := Header{
		Type:      msg.GetType(),
		Timestamp: uint64(time.Now().Unix()),
		Encrypted: msg.shouldEncrypt(),
	}
	headerBytes := header.serialize()
	payloadBytes, err := getSerializedPayload(msg)
	if err != nil {
		return nil, err
	}
	if msg.shouldEncrypt() {
		payloadBytes, err = encrypt(payloadBytes)
		if err != nil {
			return nil, err
		}
	}
	msg.setTimestamp(header.Timestamp)
	shortHash, err := getShortHash(msg)
	if err != nil {
		return nil, err
	}
	sigBytes, err := sign(shortHash)
	if err != nil {
		return nil, err
	}

	res := make([]byte, 0, len(headerBytes)+len(sigBytes)+len(payloadBytes))
	res = append(res, headerBytes...)
	res = append(res, sigBytes...)
	res = append(res, payloadBytes...)
	res = append(res, msg.serializationExtraData()...)
	return res, nil
}

type DecryptFunc = func([]byte) ([]byte, error)
type VerifySignFunc = func([]byte, []byte) bool

func Deserialize(b []byte, decrypt DecryptFunc, verifySign VerifySignFunc) (Message, error) {
	payloadIdx := HEADER_LEN + SIG_LEN
	headerB := b[:HEADER_LEN]
	sigB := b[HEADER_LEN:payloadIdx]
	payloadB := b[payloadIdx:]

	header := deserializeHeader(headerB)
	msg, err := newByType(header.Type)
	if err != nil {
		return nil, err
	}

	msg.setSignature(sigB)
	msg.setTimestamp(header.Timestamp)

	if header.Encrypted {
		payloadB, err = decrypt(payloadB)
		if err != nil {
			return nil, err
		}
	}

	err = deserializePayload(payloadB, msg)
	if err != nil {
		return nil, err
	}
	shortHash, err := getShortHash(msg)
	if err != nil {
		return nil, err
	}
	if !verifySign(shortHash, sigB) {
		return nil, fmt.Errorf("incorrect signature")
	}
	return msg, nil
}

func getShortHash(msg Message) ([]byte, error) {
	headerBytes, err := cbor.Serialize([]interface{}{msg.GetType(), msg.getTimestamp()})
	if err != nil {
		return nil, err
	}
	payloadBytes, err := getSerializedPayload(msg)
	if err != nil {
		return nil, err
	}
	data := make([]byte, 0, len(headerBytes)+len(payloadBytes))
	data = append(data, headerBytes...)
	data = append(data, payloadBytes...)
	hash := sha1.Sum(data)
	return hash[:], nil
}
