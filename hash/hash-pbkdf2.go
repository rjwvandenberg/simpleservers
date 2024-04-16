package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

const usage = `
Warning! This program has not been scrutinized.
PBKDF2 is an algorithm to derive a cryptographic key from a password and salt.
  https://en.wikipedia.org/wiki/Password_policy#NIST_guidelines
  https://pages.nist.gov/800-63-3/
  https://en.wikipedia.org/wiki/Key_derivation_function
This implementation uses HMAC-SHA3-512 as its hashing function. 

go hash-pbkdf2.go <password> <salt> <iterations>
    Supported encodings for password, salt and output key:
	- base64 (default)
`

func main() {
	if len(os.Args) != 4 {
		usageAndExit("error: invalid number of arguments")
	}

	iter, err := strconv.Atoi(os.Args[3])
	if err != nil {
		usageAndExit(fmt.Sprintf("error: iterations('%v') could not be parsed as integer", os.Args[2]))
	}

	key, err := DeriveKeyBase64(os.Args[1], os.Args[2], iter)
	if err != nil {
		fmt.Printf("%v\n", err)
		usageAndExit("error: could not derive key")
	}

	log.Println(key)
}

func DeriveKey(password []byte, salt []byte, iter int) ([]byte, error) {
	if iter < 1 {
		return nil, errors.New("iter less than 1")
	}
	return pbkdf2.Key(password, salt, iter, 64, sha3.New512), nil
}

func DeriveKeyBase64(password string, salt string, iter int) (string, error) {
	passwordDecoded, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return "", errors.Join(errors.New("error: password could not be parsed as base64"), err)
	}
	saltDecoded, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", errors.Join(errors.New("error: salt could not be parsed as base64"), err)
	}
	key, err := DeriveKey(passwordDecoded, saltDecoded, iter)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.Strict().EncodeToString(key), nil
}

func usageAndExit(errorMessage string) {
	log.Println(errorMessage)
	log.Println(usage)
	os.Exit(1)
}
