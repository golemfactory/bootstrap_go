package python

import "reflect"

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
	NatType      string        `pyobj:"nat_type"`
}

func (self *Node) ToDict() map[interface{}]interface{} {
	m := map[interface{}]interface{}{}
	v := reflect.ValueOf(self).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		tag := field.Tag.Get("pyobj")
		if tag != "" {
			m[tag] = v.Field(i).Interface()
		}
	}
	return m
}

func NodeToDict(m map[interface{}]interface{}) *Node {
	res := &Node{}
	elem := reflect.ValueOf(res).Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Type().Field(i)
		val := elem.Field(i)
		tag := field.Tag.Get("pyobj")
		if tag != "" {
			if vv, ok := m[tag]; ok && vv != nil {
				val.Set(reflect.ValueOf(vv))
			}
		}
	}
	return res
}
