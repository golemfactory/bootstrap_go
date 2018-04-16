package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"math/bits"

	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type PrivateKey struct {
	key *ecies.PrivateKey
}

func (self *PrivateKey) GetPublicKey() PublicKey {
	return PublicKey{
		&self.key.PublicKey,
	}
}

type PublicKey struct {
	key *ecies.PublicKey
}

func (self *PublicKey) Hex() string {
	return fmt.Sprintf("%064s%064s", self.key.X.Text(16), self.key.Y.Text(16))
}

func PublicKeyFromBytes(b []byte) (key PublicKey, err error) {
	if b[0] != 4 {
		err = fmt.Errorf("key not in uncompressed format, first byte=%d", b[0])
		return
	}
	b = b[1:]
	key.key = &ecies.PublicKey{
		X:     new(big.Int).SetBytes(b[:32]),
		Y:     new(big.Int).SetBytes(b[32:]),
		Curve: secp256k1.S256(),
	}
	return
}

func GeneratePrivateKey() (key PrivateKey, err error) {
	key.key, err = ecies.GenerateKey(rand.Reader, secp256k1.S256(), nil)
	return
}

// GenerateDifficultKey generates key with required difficulty.
// It should take ~1-2s for difficulty 14.
func GenerateDifficultKey(difficulty uint) (key PrivateKey, err error) {
	if difficulty > 256 {
		err = errors.New("difficulty too high")
		return
	}
	for {
		key, err = GeneratePrivateKey()
		if err != nil {
			return
		}

		if getKeyDifficulty(&key.key.PublicKey) >= int(difficulty) {
			return
		}
	}
}

func GetKeyDifficulty(key PublicKey) int {
	return getKeyDifficulty(key.key)
}

func getKeyDifficulty(key *ecies.PublicKey) int {
	pubKeyBytes := make([]byte, 0, 64)
	pubKeyBytes = append(pubKeyBytes, key.X.Bytes()...)
	pubKeyBytes = append(pubKeyBytes, key.Y.Bytes()...)
	hash := sha256.Sum256(pubKeyBytes)

	for i := 0; i < len(pubKeyBytes); i++ {
		if hash[i] != 0 {
			return i*8 + bits.LeadingZeros8(hash[i])
		}
	}

	return 8 * len(pubKeyBytes)
}
