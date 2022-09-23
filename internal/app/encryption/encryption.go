package encryption

import (
	"crypto/rand"
	_ "embed" //nolint:revive
	"io"

	"github.com/checkmarxDev/ast-sast-export/pkg/aesctr"
)

// EncryptSymmetric uses AES-CRT with HMAC inside using single key for both
func EncryptSymmetric(in io.Reader, out io.Writer, key []byte) error {
	return aesctr.Encrypt(in, out, key, key)
}

// DecryptSymmetric uses AES-CRT with HMAC inside using single key for both
func DecryptSymmetric(in io.Reader, out io.Writer, key []byte) error {
	return aesctr.Decrypt(in, out, key, key)
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
