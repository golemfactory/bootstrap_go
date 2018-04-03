package crypto

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/ishbir/elliptic"
)

func TestEncryptDecrypt(t *testing.T) {
	rand.Seed(42)
	privKey1, err := elliptic.GeneratePrivateKey(elliptic.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}
	privKey2, err := elliptic.GeneratePrivateKey(elliptic.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("asdf")
	encrytped, err := EncryptPython(privKey1, data, &privKey2.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := DecryptPython(privKey2, encrytped)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, decrypted) {
		t.Fatalf("Expected %v, got %v", data, decrypted)
	}
}

func TestKeyDifficulty(t *testing.T) {
	_, err := GenerateDifficultKey(257)
	if err.Error() != "difficulty too high" {
		t.Error("Wrong error message:", err)
	}

	curve := elliptic.Curve(714)
	X := []byte{123, 245, 242, 41, 21, 222, 148, 113, 104, 224, 38, 231, 236, 156, 161, 137, 220, 87, 120, 8, 85, 3, 173, 141, 59, 7, 254, 37, 212, 243, 147, 212}
	Y := []byte{98, 229, 253, 135, 199, 126, 195, 158, 176, 19, 177, 252, 201, 123, 12, 142, 181, 132, 99, 237, 195, 54, 196, 66, 116, 133, 166, 248, 7, 70, 216, 252}

	pubKey := &elliptic.PublicKey{
		Curve: curve,
		X:     X,
		Y:     Y,
	}

	if GetKeyDifficulty(pubKey) == 14 {
		t.Error("Key should be difficult", pubKey)
	}
}
