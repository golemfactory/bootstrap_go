package cbor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCbor(t *testing.T) {
	type TestData struct {
		Int    int
		String string
		Map    map[string]int
	}
	data := TestData{
		Int:    1,
		String: "string",
		Map:    map[string]int{"key": 1},
	}
	serialized, err := Serialize(data)
	require.NoError(t, err)
	var deserialized TestData
	err = Deserialize(serialized, &deserialized)
	require.NoError(t, err)
	assert.Equal(t, data, deserialized)
}
