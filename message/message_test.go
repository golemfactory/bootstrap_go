package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testImpl(t *testing.T, msg Message) Message {
	encryptCalled := false
	encryptFunc := func(data []byte) ([]byte, error) {
		encryptCalled = true
		return data, nil
	}

	decryptCalled := false
	decryptFunc := func(data []byte) ([]byte, error) {
		decryptCalled = true
		return data, nil
	}

	sig := make([]byte, SIG_LEN)
	for i := 0; i < SIG_LEN; i++ {
		sig[i] = byte(i)
	}
	signCalled := false
	signFunc := func(msg Message) {
		signCalled = true
		msg.GetBaseMessage().Sig = sig
	}

	serialized, err := Serialize(msg, encryptFunc, signFunc)
	require.NoError(t, err)
	assert.Equal(t, msg.ShouldEncrypt(), encryptCalled)
	assert.True(t, signCalled)

	deserialized, err := Deserialize(serialized, decryptFunc)
	require.NoError(t, err)
	assert.Equal(t, msg.ShouldEncrypt(), decryptCalled)

	baseMsg := deserialized.GetBaseMessage()
	assert.Equal(t, msg.GetType(), baseMsg.Header.Type)
	assert.Equal(t, msg.ShouldEncrypt(), baseMsg.Header.Encrypted)
	assert.Equal(t, sig, baseMsg.Sig)
	return deserialized
}

func TestSerializeationEncrypted(t *testing.T) {
	const RAND_VAL = 0.1337
	msg := &RandVal{
		RandVal: RAND_VAL,
	}
	require.True(t, msg.ShouldEncrypt())
	deserialized := testImpl(t, msg)

	castedMsg, ok := deserialized.(*RandVal)
	require.True(t, ok)
	assert.Equal(t, RAND_VAL, castedMsg.RandVal)
}

func TestSerializeationNotEncrypted(t *testing.T) {
	const REASON = "Unittest"
	msg := &Disconnect{
		Reason: REASON,
	}
	require.False(t, msg.ShouldEncrypt())
	deserialized := testImpl(t, msg)

	castedMsg, ok := deserialized.(*Disconnect)
	require.True(t, ok)
	assert.Equal(t, REASON, castedMsg.Reason)
}
