package main

import (
	"encoding/hex"
	"testing"
)

// Testing data
// openssl kdf -keylen 64 -kdfopt digest:SHA3-512 -kdfopt pass:A -kdfopt salt:B -kdfopt iter:1 PBKDF2 | tr -d : | tr '[:upper:]' '[:lower:]'
func TestHash(t *testing.T) {
	expected := "f0e937a862d46c987932c9b1c11a0889e0f99938824ecc1b4dfc5d40a63245371b2b220e28e5e0562ac3f8bd0b5f1751c8dd846ea44640aca4f2a193f488dbfc"
	key, err := DeriveKey([]byte{'A'}, []byte{'B'}, 1)
	if err != nil {
		t.Fatalf("error in test: %v", err)
	}
	hexkey := hex.EncodeToString(key)
	if expected != hexkey {
		t.Fatalf("key mismatch: \nkey : %v\nwant: %v\n", hexkey, expected)
	}
}

func TestHashManyIterations(t *testing.T) {
	expected := "b5b0168c663468f4f09ff61094f140f27f703d693d32850cb8f9a357a6ed6c8be66c80610f6497cab4b931180cd31a9d1b220c34c00b59a11d609696e8819e6e"
	key, err := DeriveKey([]byte{'A'}, []byte{'B'}, 10000)
	if err != nil {
		t.Fatalf("error in test: %v", err)
	}
	hexkey := hex.EncodeToString(key)
	if expected != hexkey {
		t.Fatalf("key mismatch: \nkey : %v\nwant: %v\n", hexkey, expected)
	}
}

func TestHashBase64(t *testing.T) {
	expected := "tbAWjGY0aPTwn/YQlPFA8n9wPWk9MoUMuPmjV6btbIvmbIBhD2SXyrS5MRgM0xqdGyIMNMALWaEdYJaW6IGebg=="
	key, err := DeriveKeyBase64("QQ==", "Qg==", 10000)
	if err != nil {
		t.Fatalf("error in test: %v", err)
	}
	if expected != key {
		t.Fatalf("key mismatch: \nkey : %v\nwant: %v\n", key, expected)
	}
}

func TestManyIterations(t *testing.T) {
	DeriveKeyBase64("QQ==", "Qg==", 600000)
}
