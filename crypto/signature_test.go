package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignature(t *testing.T) {
	key, err := GeneratePrivateKey()
	data := []byte("asdfasdfasdfasdfasdfasdfasdfasdf")
	sig, err := key.Sign(data)
	require.NoError(t, err)
	pub := key.GetPublicKey()
	ok := pub.VerifySign(data, sig)
	assert.True(t, ok)
}
