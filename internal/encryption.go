package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
)

var buildTimeRSAPublicKey string

var RSAPublicKey = fmt.Sprintf(`
-----BEGIN PUBLIC KEY-----
%s
-----END PUBLIC KEY-----
`, buildTimeRSAPublicKey)

func RSAEncrypt(key, plaintext []byte) ([]byte, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return []byte{}, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return []byte{}, err
	}

	publicKey, ok := pub.(*rsa.PublicKey)

	if !ok {
		return []byte{}, fmt.Errorf("invalid public key")
	}

	label := []byte("")

	// crypto/rand.Reader is a good source of entropy for randomizing the
	// encryption function.
	rng := rand.Reader

	return rsa.EncryptOAEP(sha256.New(), rng, publicKey, plaintext, label)
}

func AESEncrypt(key, plaintext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return []byte{}, err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return []byte{}, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

func CreateSymmetricKey(length int) ([]byte, error) {
	key := make([]byte, length)

	_, err := rand.Read(key)
	if err != nil {
		return []byte{}, err
	}

	return key, nil
}
