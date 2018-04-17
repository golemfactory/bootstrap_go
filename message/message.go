package message

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"reflect"
	"time"

	"github.com/golemfactory/bootstrap_go/cbor"
)

type Message interface {
	GetSignature() []byte
	GetType() uint16

	shouldEncrypt() bool
	getTimestamp() uint64
	setTimestamp(ts uint64)
	setSignature(sig []byte)
	serializationExtraData() []byte
}

type BaseMessage struct {
	Timestamp uint64
	Sig       []byte
}

func (self *BaseMessage) GetSignature() []byte {
	return self.Sig
}

func (self *BaseMessage) getTimestamp() uint64 {
	return self.Timestamp
}

func (self *BaseMessage) setTimestamp(ts uint64) {
	self.Timestamp = ts
}

func (self *BaseMessage) setSignature(sig []byte) {
	self.Sig = sig
}

func (self *BaseMessage) serializationExtraData() []byte {
	return []byte{}
}

func getShortHash(msg Message) []byte {
	payload := GetPayload(msg)
	data := make([]byte, 0)
	headerBytes, _ := cbor.Serialize([]interface{}{msg.GetType(), msg.getTimestamp()})
	payloadBytes, _ := serializePayload(payload)
	data = append(data, headerBytes...)
	data = append(data, payloadBytes...)
	hash := sha1.Sum(data)
	return hash[:]
}

// slot is a pair fo python field's name and value
type MessageSlot = []interface{}

// list of MessageSlots
type MessagePayload = []interface{}

func GetPayload(msg Message) MessagePayload {
	payload := MessagePayload{}
	v := reflect.ValueOf(msg).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		val := v.Field(i)
		tag := field.Tag.Get("msg_slot")
		if tag != "" {
			payload = append(payload, MessageSlot{tag, val.Interface()})
		}
	}
	return payload
}

const (
	HEADER_LEN = 11
	SIG_LEN    = 65
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

type DecryptFunc = func([]byte) ([]byte, error)
type VerifySignFunc = func([]byte, []byte) bool

func Deserialize(b []byte, decrypt DecryptFunc) (Message, error) {
	payloadIdx := HEADER_LEN + SIG_LEN
	headerB := b[:HEADER_LEN]
	sigB := b[HEADER_LEN:payloadIdx]
	payloadB := b[payloadIdx:]

	header := deserializeHeader(headerB)
	var msg Message
	if header.Type == MSG_HELLO_TYPE {
		msg = &Hello{}
	} else if header.Type == MSG_RAND_VAL_TYPE {
		msg = &RandVal{}
	} else if header.Type == MSG_DISCONNECT_TYPE {
		msg = &Disconnect{}
	} else if header.Type == MSG_PEERS_TYPE {
		msg = &Peers{}
	} else {
		return nil, fmt.Errorf("unsupported msg type %d", header.Type)
	}

	msg.setSignature(sigB)
	msg.setTimestamp(header.Timestamp)

	var err error
	if header.Encrypted {
		payloadB, err = decrypt(payloadB)
		if err != nil {
			return nil, err
		}
	}

	var maybeSlots interface{}
	err = cbor.Deserialize(payloadB, &maybeSlots)
	if err != nil {
		return nil, err
	}
	slots, ok := maybeSlots.(MessagePayload)
	if !ok {
		return nil, fmt.Errorf("incorrect format of message payload")
	}
	deserializePayload(slots, msg)
	return msg, nil
}

func deserializePayload(slotsList MessagePayload, msg Message) {
	slots := make(map[string]interface{})
	for _, s := range slotsList {
		slot, ok := s.(MessageSlot)
		if !ok {
			fmt.Printf("Couldn't cast slot %+v\n", s)
			continue
		}
		if slotName, ok := slot[0].(string); ok {
			slots[slotName] = slot[1]
		} else {
			fmt.Printf("Expected slot name to be a string, got %+v", slot[0])
		}
	}

	v := reflect.ValueOf(msg).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		val := v.Field(i)
		tag := field.Tag.Get("msg_slot")
		if tag != "" {
			if vv, ok := slots[tag]; ok && vv != nil {
				val.Set(reflect.ValueOf(vv))
			}
		}
	}
}

func serializePayload(payload MessagePayload) ([]byte, error) {
	return cbor.Serialize(payload)
}

type EncryptFunc = func([]byte) ([]byte, error)
type SignFunc = func([]byte) ([]byte, error)

func Serialize(msg Message, encrypt EncryptFunc, sign SignFunc) ([]byte, error) {
	header := Header{}
	header.Type = msg.GetType()
	header.Timestamp = uint64(time.Now().Unix())
	header.Encrypted = msg.shouldEncrypt()
	headerBytes := header.serialize()
	payloadBytes, err := serializePayload(GetPayload(msg))
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
	sigBytes, err := sign(getShortHash(msg))
	if err != nil {
		return nil, err
	}

	res := make([]byte, 0)
	res = append(res, headerBytes...)
	res = append(res, sigBytes...)
	res = append(res, payloadBytes...)
	res = append(res, msg.serializationExtraData()...)
	return res, nil
}

const (
	MSG_HELLO_TYPE      = 0
	MSG_RAND_VAL_TYPE   = 1
	MSG_DISCONNECT_TYPE = 2
	MSG_PEERS_TYPE      = 1004
)

type Hello struct {
	BaseMessage
	RandVal              float64                     `msg_slot:"rand_val"`
	ProtoId              string                      `msg_slot:"proto_id"`
	NodeName             string                      `msg_slot:"node_name"`
	NodeInfo             map[interface{}]interface{} `msg_slot:"node_info"`
	Port                 uint64                      `msg_slot:"port"`
	ClientVer            string                      `msg_slot:"client_ver"`
	ClientKeyId          string                      `msg_slot:"client_key_id"`
	SolveChallange       bool                        `msg_slot:"solve_challenge"`
	Challange            interface{}                 `msg_slot:"challenge"`
	Difficulty           uint64                      `msg_slot:"difficulty"`
	Metadata             interface{}                 `msg_slot:"metadata"`
	GolemMessagesVersion string                      `msg_slot:"_version"`
}

func (self *Hello) GetType() uint16 {
	return MSG_HELLO_TYPE
}

func (self *Hello) shouldEncrypt() bool {
	return false
}

func (self *Hello) serializationExtraData() []byte {
	res := make([]byte, 0, 32)
	vlen := len(self.GolemMessagesVersion)
	res = append(res, byte(vlen))
	res = append(res, []byte(self.GolemMessagesVersion)...)
	res = append(res, make([]byte, 31-vlen)...)
	return res
}

type RandVal struct {
	BaseMessage
	RandVal float64 `msg_slot:"rand_val"`
}

func (self *RandVal) GetType() uint16 {
	return MSG_RAND_VAL_TYPE
}

func (self *RandVal) shouldEncrypt() bool {
	return true
}

type DisconnectReason = string

const (
	DISCONNECT_PROTOCOL_VERSION DisconnectReason = "protocol_version"
	DISCONNECT_UNVERIFIED       DisconnectReason = "unverified"
	DISCONNECT_BOOTSTRAP        DisconnectReason = "bootstrap"
	DISCONNECT_KEY_DIFFICULTY   DisconnectReason = "key_not_difficult"
)

type Disconnect struct {
	BaseMessage
	Reason DisconnectReason `msg_slot:"reason"`
}

func (self *Disconnect) GetType() uint16 {
	return MSG_DISCONNECT_TYPE
}

func (self *Disconnect) shouldEncrypt() bool {
	return false
}

type Peers struct {
	BaseMessage
	Peers []interface{} `msg_slot:"peers"`
}

func (self *Peers) GetType() uint16 {
	return MSG_PEERS_TYPE
}

func (self *Peers) shouldEncrypt() bool {
	return true
}
