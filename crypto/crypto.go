package crypto

import (
	"crypto/rand"

	"github.com/golemfactory/bootstrap_go/crypto/ecies"
)

func Encrypt(data []byte, pubKey PublicKey) ([]byte, error) {
	return ecies.Encrypt(rand.Reader, pubKey.key, data, nil, nil)
}

func (self *PrivateKey) Decrypt(raw []byte) ([]byte, error) {
	return ecies.Decrypt(self.key, raw, nil, nil)
}
