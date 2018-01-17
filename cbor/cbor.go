package cbor

import (
	"bufio"
	"bytes"

	impl "github.com/whyrusleeping/cbor/go"
)

func Serialize(obj interface{}) ([]byte, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	encoder := impl.NewEncoder(writer)
	err := encoder.Encode(obj)
	if err != nil {
		return nil, err
	}
	writer.Flush()
	return b.Bytes(), nil
}

func Deserialize(input []byte, obj interface{}) error {
	reader := bytes.NewReader(input)
	decoder := impl.NewDecoder(reader)
	var pyObjectDecoder PyObjectDecoder
	decoder.TagDecoders[pyObjectDecoder.GetTag()] = &pyObjectDecoder
	return decoder.Decode(obj)
}
