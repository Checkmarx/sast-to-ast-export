package internal

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

var buildTimeRSAPublicKey string

var RSAPublicKey = fmt.Sprintf(`
-----BEGIN PUBLIC KEY-----
%s
-----END PUBLIC KEY-----
`, buildTimeRSAPublicKey)

func Encrypt(key string, plaintext string) (string, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	publicKey, ok := pub.(*rsa.PublicKey)

	if !ok {
		return "", fmt.Errorf("invalid public key")
	}

	label := []byte("")

	// crypto/rand.Reader is a good source of entropy for randomizing the
	// encryption function.
	rng := rand.Reader

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, publicKey, []byte(plaintext), label)
	if err != nil {
		return "", err
	}
	base64Cyphertext := bytes.NewBufferString("")
	encoder := base64.NewEncoder(base64.StdEncoding, base64Cyphertext)
	encoder.Write(ciphertext)
	encoder.Close()

	return base64Cyphertext.String(), nil
}
