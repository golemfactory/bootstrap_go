package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerialization(t *testing.T) {
	header := Header{
		Type:      134,
		Timestamp: 434343,
		Encrypted: true,
	}
	serialized := header.serialize()
	deserialized := deserializeHeader(serialized)
	assert.Equal(t, header, deserialized)
}
