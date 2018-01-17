package cbor

import (
	"fmt"
	"io"
	"reflect"

	impl "github.com/whyrusleeping/cbor/go"
)

const CBOR_TAG = 0xef

type PyObject interface {
	GetPyObjectName() string
}

func ToCBOR(obj PyObject, w io.Writer, enc *impl.Encoder) error {
	_, err := w.Write([]byte{0xd8, CBOR_TAG})
	if err != nil {
		return err
	}
	m := map[string]interface{}{}
	m["py/object"] = obj.GetPyObjectName()
	v := reflect.ValueOf(obj).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		tag := field.Tag.Get("pyobj")
		if tag != "" {
			m[tag] = v.Field(i).Interface()
		}
	}
	return enc.Encode(m)
}

var pythonTypes = make(map[string]func() interface{})

func RegisterPythonType(name string, factory func() interface{}) {
	pythonTypes[name] = factory
}

type PyObjectDecoder struct{}

func (self *PyObjectDecoder) GetTag() uint64 {
	return CBOR_TAG
}

func (self *PyObjectDecoder) DecodeTarget() interface{} {
	return make(map[string]interface{})
}

func (self *PyObjectDecoder) PostDecode(v interface{}) (interface{}, error) {
	m := v.(map[string]interface{})
	var res interface{}
	pythonType, ok := m["py/object"].(string)
	if !ok {
		return nil, fmt.Errorf("missing py/object")
	}
	if factory, ok := pythonTypes[pythonType]; ok {
		res = factory()
	} else {
		return nil, fmt.Errorf("unsupported py/object %s", m["py/object"])
	}

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
	return res, nil
}
