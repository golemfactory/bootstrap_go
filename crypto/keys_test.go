package crypto

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

func TestKeyDifficulty(t *testing.T) {
	_, err := GenerateDifficultKey(257)
	if err.Error() != "difficulty too high" {
		t.Error("Wrong error message:", err)
	}

	curve := secp256k1.S256()
	X := []byte{123, 245, 242, 41, 21, 222, 148, 113, 104, 224, 38, 231, 236, 156, 161, 137, 220, 87, 120, 8, 85, 3, 173, 141, 59, 7, 254, 37, 212, 243, 147, 212}
	Y := []byte{98, 229, 253, 135, 199, 126, 195, 158, 176, 19, 177, 252, 201, 123, 12, 142, 181, 132, 99, 237, 195, 54, 196, 66, 116, 133, 166, 248, 7, 70, 216, 252}

	pubKey := &ecies.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(X),
		Y:     new(big.Int).SetBytes(Y),
	}

	if getKeyDifficulty(pubKey) == 14 {
		t.Error("Key should be difficult", pubKey)
	}
}
