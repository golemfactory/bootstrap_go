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
	return decoder.Decode(obj)
}
