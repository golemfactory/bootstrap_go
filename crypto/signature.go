package crypto

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

func (self *PrivateKey) Sign(data []byte) ([]byte, error) {
	return secp256k1.Sign(data, self.key.Key)
}

func (self *PublicKey) VerifySign(data []byte, signature []byte) (bool, error) {
	keyBytes := []byte{0x04}
	keyBytes = append(keyBytes, self.key.X...)
	keyBytes = append(keyBytes, self.key.Y...)
	recoveredKey, err := secp256k1.RecoverPubkey(data, signature)
	if err != nil {
		return false, fmt.Errorf("unable to recover public key: %v", err)
	}
	return bytes.Equal(recoveredKey, keyBytes), nil
}
