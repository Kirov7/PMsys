package encrypts

import (
	"testing"
)

const AESKey = "asdfghjklzxcvbnmqwertyui"

func TestEncode(t *testing.T) {
	enc, err := EncryptInt64(45, AESKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(enc)
}

func TestDecode(t *testing.T) {
	dec, err := Decrypt("7adb", AESKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dec)
}
