package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	require.NoError(t, err)
	data := []byte("asdf")
	encrypted, err := Encrypt(data, privKey.GetPublicKey())
	require.NoError(t, err)
	decrypted, err := privKey.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}
