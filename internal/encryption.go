package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"io"
)

// buildTimeRSAPublicKey is a base64-encoded RSA public key, defined at build time
var buildTimeRSAPublicKey string

// CreatePublicKeyFromKeyBytes creates an RSA PublicKey structure from public key bytes
func CreatePublicKeyFromKeyBytes(keyBytes []byte) (*rsa.PublicKey, error) {
	pub, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := pub.(*rsa.PublicKey)

	if !ok {
		return nil, fmt.Errorf("invalid public key")
	}

	return publicKey, nil
}

// EncryptAsymmetric does RSA-OAEP with SHA-256
func EncryptAsymmetric(key *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, key, plaintext, []byte{})
}

// EncryptSymmetric does AES-GCM
func EncryptSymmetric(key, plaintext []byte) ([]byte, error) {
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

// CreateSymmetricKey creates a cryptographically secure random key of the specified length
func CreateSymmetricKey(length int) ([]byte, error) {
	key := make([]byte, length)

	_, err := rand.Read(key)
	if err != nil {
		return []byte{}, err
	}

	return key, nil
}
