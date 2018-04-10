package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("asdf")
	encrypted, err := Encrypt(data, privKey.GetPublicKey())
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := privKey.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, decrypted) {
		t.Fatalf("Expected %v, got %v", data, decrypted)
	}
}
