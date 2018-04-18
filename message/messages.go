package message

import "fmt"

const (
	MSG_HELLO_TYPE      = 0
	MSG_RAND_VAL_TYPE   = 1
	MSG_DISCONNECT_TYPE = 2
	MSG_PEERS_TYPE      = 1004
)

type Hello struct {
	baseMessage
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
	baseMessage
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
	baseMessage
	Reason DisconnectReason `msg_slot:"reason"`
}

func (self *Disconnect) GetType() uint16 {
	return MSG_DISCONNECT_TYPE
}

func (self *Disconnect) shouldEncrypt() bool {
	return false
}

type Peers struct {
	baseMessage
	Peers []interface{} `msg_slot:"peers"`
}

func (self *Peers) GetType() uint16 {
	return MSG_PEERS_TYPE
}

func (self *Peers) shouldEncrypt() bool {
	return true
}

var registeredTypes = make(map[uint16]func() Message)

func newByType(typ uint16) (Message, error) {
	factory, ok := registeredTypes[typ]
	if !ok {
		return nil, fmt.Errorf("unsupported msg type %d", typ)
	}
	return factory(), nil
}

func init() {
	factories := []func() Message{
		func() Message { return &Hello{} },
		func() Message { return &RandVal{} },
		func() Message { return &Disconnect{} },
		func() Message { return &Peers{} },
	}
	for _, factory := range factories {
		registeredTypes[factory().GetType()] = factory
	}
	if len(factories) != len(registeredTypes) {
		panic("Duplicated message types")
	}
}
