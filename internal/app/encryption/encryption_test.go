package encryption

import (
	"bytes"
	"testing"

	"github.com/checkmarxDev/ast-sast-export/pkg/aesctr"
	"github.com/stretchr/testify/assert"
)

const (
	plaintext          = "this is a test"
	symmetricKeyLength = 32
)

func TestEncryptSymmetric(t *testing.T) {
	key, keyErr := CreateSymmetricKey(symmetricKeyLength)
	assert.NoError(t, keyErr)

	// encrypt
	plain := bytes.NewReader([]byte(plaintext))
	enc := bytes.NewBuffer([]byte{})
	err := EncryptSymmetric(plain, enc, key)
	assert.NoError(t, err)

	// decrypt
	decr := bytes.NewBuffer([]byte{})
	err = aesctr.Decrypt(enc, decr, key, key)
	assert.NoError(t, err)

	// check decrypted matches plaintext
	assert.Equal(t, plaintext, decr.String())
}

func TestCreateSymmetricKey(t *testing.T) {
	result1, err1 := CreateSymmetricKey(symmetricKeyLength)
	result2, err2 := CreateSymmetricKey(symmetricKeyLength)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, result1, result2)
	assert.Equal(t, symmetricKeyLength, len(result1))
	assert.Equal(t, symmetricKeyLength, len(result2))
}
