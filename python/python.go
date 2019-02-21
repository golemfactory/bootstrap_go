package python

import (
	"fmt"
	"reflect"
)

type serializableToDict interface {
	ToDict() map[interface{}]interface{}
}

func toDict(obj interface{}) map[interface{}]interface{} {
	m := map[interface{}]interface{}{}
	v := reflect.ValueOf(obj).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		tag := field.Tag.Get("pyobj")
		if tag != "" {
			val := v.Field(i).Interface()
			if serializableVal, ok := val.(serializableToDict); ok {
				val = serializableVal.ToDict()
			}
			m[tag] = val
		}
	}
	return m
}

type Node struct {
	NodeName     string        `pyobj:"node_name"`
	Key          string        `pyobj:"key"`
	PrvPort      uint64        `pyobj:"prv_port"`
	PubPort      uint64        `pyobj:"pub_port"`
	P2pPrvPort   uint64        `pyobj:"p2p_prv_port"`
	P2pPubPort   uint64        `pyobj:"p2p_pub_port"`
	PrvAddr      string        `pyobj:"prv_addr"`
	PubAddr      string        `pyobj:"pub_addr"`
	PrvAddresses []interface{} `pyobj:"prv_addresses"`
	NatType      []interface{} `pyobj:"nat_type"`
}

func (self *Node) ToDict() map[interface{}]interface{} {
	return toDict(self)
}

func DictToNode(m map[interface{}]interface{}) (*Node, error) {
	res := &Node{}
	elem := reflect.ValueOf(res).Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Type().Field(i)
		val := elem.Field(i)
		tag := field.Tag.Get("pyobj")
		if tag != "" {
			if vv, ok := m[tag]; ok && vv != nil {
				if !reflect.TypeOf(vv).AssignableTo(val.Type()) {
					return nil, fmt.Errorf("can't assign %v to %v for property %v", reflect.TypeOf(vv), val.Type(), tag)
				}
				val.Set(reflect.ValueOf(vv))
			}
		}
	}
	return res, nil
}

type Peer struct {
	Address  string `pyobj:"address"`
	Port     uint64 `pyobj:"port"`
	Node     *Node  `pyobj:"node"`
	NodeName string `pyobj:"node_name"`
}

func (self *Peer) ToDict() map[interface{}]interface{} {
	return toDict(self)
}
