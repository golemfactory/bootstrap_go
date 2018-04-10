package crypto

import (
	"crypto/rand"

	"github.com/ethereum/go-ethereum/crypto/ecies"
)

func Encrypt(data []byte, pubKey PublicKey) ([]byte, error) {
	return ecies.Encrypt(rand.Reader, pubKey.key, data, nil, nil)
}

func (self *PrivateKey) Decrypt(raw []byte) ([]byte, error) {
	return self.key.Decrypt(rand.Reader, raw, nil, nil)
}
