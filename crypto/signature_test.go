package crypto

import "testing"

func TestSignature(t *testing.T) {
	key, err := GeneratePrivateKey()
	data := []byte("asdfasdfasdfasdfasdfasdfasdfasdf")
	sig, err := key.Sign(data)
	if err != nil {
		t.Fatal(err)
	}
	pub := key.GetPublicKey()
	ok := pub.VerifySign(data, sig)
	if !ok {
		t.Fatal("incorrect signature", err)
	}
}
