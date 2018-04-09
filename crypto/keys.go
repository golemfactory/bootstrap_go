package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/bits"

	"github.com/ishbir/elliptic"
)

type PrivateKey struct {
	key *elliptic.PrivateKey
}

type PublicKey struct {
	key *elliptic.PublicKey
}

func (self *PrivateKey) GetPublicKey() PublicKey {
	return PublicKey{
		&self.key.PublicKey,
	}
}

func (self *PrivateKey) GetPubKeyHex() string {
	return hex.EncodeToString(self.key.PublicKey.X) + hex.EncodeToString(self.key.PublicKey.Y)
}

func PublicKeyFromBytes(b []byte) (key PublicKey, err error) {
	key.key, err = elliptic.PublicKeyFromUncompressedBytes(elliptic.Secp256k1, b)
	return
}

func GeneratePrivateKey() (key PrivateKey, err error) {
	key.key, err = elliptic.GeneratePrivateKey(elliptic.Secp256k1)
	return
}

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

func getKeyDifficulty(key *elliptic.PublicKey) int {
	pubKeyBytes := make([]byte, 0, 32)
	pubKeyBytes = append(pubKeyBytes, key.X...)
	pubKeyBytes = append(pubKeyBytes, key.Y...)
	hash := sha256.Sum256(pubKeyBytes)

	for i := 0; i < len(pubKeyBytes); i++ {
		if hash[i] != 0 {
			return i*8 + bits.LeadingZeros8(hash[i])
		}
	}

	return 8 * len(pubKeyBytes)
}
