package crypto

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyDifficulty(t *testing.T) {
	_, err := GenerateDifficultKey(257)
	assert.Error(t, err)

	curve := secp256k1.S256()
	X := []byte{123, 245, 242, 41, 21, 222, 148, 113, 104, 224, 38, 231, 236, 156, 161, 137, 220, 87, 120, 8, 85, 3, 173, 141, 59, 7, 254, 37, 212, 243, 147, 212}
	Y := []byte{98, 229, 253, 135, 199, 126, 195, 158, 176, 19, 177, 252, 201, 123, 12, 142, 181, 132, 99, 237, 195, 54, 196, 66, 116, 133, 166, 248, 7, 70, 216, 252}

	pubKey := &ecies.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(X),
		Y:     new(big.Int).SetBytes(Y),
	}

	assert.Equal(t, 18, getKeyDifficulty(pubKey))
}

func TestPublicKeyHex(t *testing.T) {
	pubKey := &ecies.PublicKey{
		Curve: secp256k1.S256(),
		X:     new(big.Int),
		Y:     new(big.Int),
	}
	_, ok := pubKey.X.SetString("413948280899852118482748229132309424713966529587595585748106469290859442165", 10)
	require.True(t, ok)
	_, ok = pubKey.Y.SetString("68586480335090537784423010349535964989539087847494459304095557215436234524160", 10)
	require.True(t, ok)
	key := PublicKey{
		key: pubKey,
	}
	assert.Equal(t, 128, len(key.Hex()))
}

func benchmarkDifficultKeyGeneration(difficulty uint, b *testing.B) {
    for n := 0; n < b.N; n++ {
         GenerateDifficultKey(difficulty)
    }
}

// $ go test  -bench=. -benchtime=1m ./crypto/
// goos: darwin
// goarch: amd64
// BenchmarkDifficultKeyGeneration14-8   	      50	2720878068 ns/op
// BenchmarkDifficultKeyGeneration16-8   	      50	7893233868 ns/op


func BenchmarkDifficultKeyGeneration14(b *testing.B) {
    benchmarkDifficultKeyGeneration(14, b)
}

func BenchmarkDifficultKeyGeneration16(b *testing.B) {
    benchmarkDifficultKeyGeneration(16, b)
}

func BenchmarkDifficultKeyGeneration18(b *testing.B) {
    benchmarkDifficultKeyGeneration(18, b)
}

func BenchmarkDifficultKeyGeneration20(b *testing.B) {
    benchmarkDifficultKeyGeneration(20, b)
}