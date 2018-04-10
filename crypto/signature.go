package crypto

import (
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

func (self *PrivateKey) Sign(data []byte) ([]byte, error) {
	return secp256k1.Sign(data, self.key.D.Bytes())
}

func (self *PublicKey) VerifySign(data []byte, signature []byte) bool {
	pubKeyBytes := make([]byte, 0, 65)
	pubKeyBytes = append(pubKeyBytes, 0x4)
	pubKeyBytes = append(pubKeyBytes, self.key.X.Bytes()...)
	pubKeyBytes = append(pubKeyBytes, self.key.Y.Bytes()...)
	return secp256k1.VerifySignature(pubKeyBytes, data, signature[:64])
}
