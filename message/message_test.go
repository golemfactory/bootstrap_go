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
	signFunc := func(data []byte) ([]byte, error) {
		signCalled = true
		return sig, nil
	}

	verifySignCalled := false
	verifySignFunc := func(data []byte, sig []byte) bool {
		verifySignCalled = true
		return true
	}

	serialized, err := Serialize(msg, encryptFunc, signFunc)
	require.NoError(t, err)
	assert.Equal(t, msg.shouldEncrypt(), encryptCalled)
	assert.True(t, signCalled)

	deserialized, err := Deserialize(serialized, decryptFunc, verifySignFunc)
	require.NoError(t, err)
	assert.Equal(t, msg.shouldEncrypt(), decryptCalled)
	assert.True(t, verifySignCalled)

	assert.Equal(t, msg.GetType(), deserialized.GetType())
	assert.Equal(t, sig, deserialized.getSignature())
	assert.NotZero(t, deserialized.getTimestamp())
	return deserialized
}

func TestSerializationEncrypted(t *testing.T) {
	const RAND_VAL = 0.1337
	msg := &RandVal{
		RandVal: RAND_VAL,
	}
	require.True(t, msg.shouldEncrypt())
	deserialized := testImpl(t, msg)

	castedMsg, ok := deserialized.(*RandVal)
	require.True(t, ok)
	assert.Equal(t, RAND_VAL, castedMsg.RandVal)
}

func TestSerializationNotEncrypted(t *testing.T) {
	const REASON = "Unittest"
	msg := &Disconnect{
		Reason: REASON,
	}
	require.False(t, msg.shouldEncrypt())
	deserialized := testImpl(t, msg)

	castedMsg, ok := deserialized.(*Disconnect)
	require.True(t, ok)
	assert.Equal(t, REASON, castedMsg.Reason)
}
