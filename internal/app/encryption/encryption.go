package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"fmt"
	"io"

	"github.com/checkmarxDev/ast-sast-export/pkg/aesctr"
)

//nolint:stylecheck
//go:embed public.key
var BuildTimeRSAPublicKey string

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

// EncryptSymmetric uses AES-CRT with HMAC inside
// using single key for both
func EncryptSymmetric(in io.Reader, out io.Writer, key []byte) error {
	return aesctr.Encrypt(in, out, key, key)
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
