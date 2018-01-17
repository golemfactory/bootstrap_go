package bootstrap

import (
	"bufio"
	"bytes"

	"github.com/whyrusleeping/cbor/go"
)

// https://pypi.python.org/pypi/cbor2/3.0.4
// Semantics: Mark shared value
type Tag28Decoder struct{}

func (self *Tag28Decoder) GetTag() uint64 {
	return 28
}

func (self *Tag28Decoder) DecodeTarget() interface{} {
	var v interface{}
	return &v
}

func (self *Tag28Decoder) PostDecode(v interface{}) (interface{}, error) {
	return *v.(*interface{}), nil
}

func cborSerialize(obj interface{}) ([]byte, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	encoder := cbor.NewEncoder(writer)
	err := encoder.Encode(obj)
	if err != nil {
		return nil, err
	}
	writer.Flush()
	return b.Bytes(), nil
}

func cborDeserialize(input []byte, obj interface{}) error {
	reader := bytes.NewReader(input)
	decoder := cbor.NewDecoder(reader)
	var pyObjectDecoder PyObjectDecoder
	decoder.TagDecoders[pyObjectDecoder.GetTag()] = &pyObjectDecoder
	var tag28Decoder Tag28Decoder
	decoder.TagDecoders[tag28Decoder.GetTag()] = &tag28Decoder
	return decoder.Decode(obj)
}
