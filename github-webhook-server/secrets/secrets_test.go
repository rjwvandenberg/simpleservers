package secrets

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"
)

func TestHMACSha256(t *testing.T) {
	secretbytes := []byte("It's a Secret to Everybody")
	hashstring := "sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"[7:]
	sh := New("", secretbytes)
	hash, _ := hex.DecodeString(hashstring)
	if !sh.Validate(hash, []byte("Hello, World!")) {
		t.Fail()
	}
}

func TestExampleA(t *testing.T) {
	testfile, _ := os.ReadFile("../files/post-test.http")
	split := bytes.Split(testfile, []byte("\n\n"))
	secretbytes := []byte("A")
	hashstring := string(bytes.Split(bytes.Split(split[0], []byte("\n"))[4], []byte(" "))[1])[7:]
	sh := New("", secretbytes)
	hash, _ := hex.DecodeString(hashstring)
	if !(sh.Validate(hash, split[1])) {
		t.Fail()
	}

}
