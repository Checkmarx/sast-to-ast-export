package internal

import (
	aes2 "crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	plaintext           = "this is a test"
	asymmetricKeyLength = 4096
	symmetricKeyLength  = 32
)

func TestCreatePublicKeyFromKeyBytes(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, asymmetricKeyLength)
	assert.NoError(t, err)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	assert.NoError(t, err)

	result, err := CreatePublicKeyFromKeyBytes(publicKeyBytes)

	assert.NoError(t, err)
	assert.Equal(t, rsaKey.PublicKey.E, result.E)
	assert.Equal(t, rsaKey.PublicKey.N, result.N)
}

func TestEncryptAsymmetric(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, asymmetricKeyLength)
	assert.NoError(t, err)

	// encrypt
	ciphertext, encryptErr := EncryptAsymmetric(&rsaKey.PublicKey, []byte(plaintext))
	assert.NoError(t, encryptErr)

	// decrypt
	result, decryptErr := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, ciphertext, []byte{})
	assert.NoError(t, decryptErr)

	// check decrypted matches plaintext
	assert.Equal(t, plaintext, string(result))
}

func TestEncryptSymmetric(t *testing.T) {
	key, keyErr := CreateSymmetricKey(symmetricKeyLength)
	assert.NoError(t, keyErr)

	// encrypt
	ciphertext, encryptErr := EncryptSymmetric(key, []byte(plaintext))
	assert.NoError(t, encryptErr)

	// decrypt
	block, blockErr := aes2.NewCipher(key)
	assert.NoError(t, blockErr)

	gcm, cipherErr := cipher.NewGCM(block)
	assert.NoError(t, cipherErr)

	nonceSize := gcm.NonceSize()
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	result, decryptErr := gcm.Open(nil, nonce, actualCiphertext, nil)
	assert.NoError(t, decryptErr)

	// check decrypted matches plaintext
	assert.Equal(t, plaintext, string(result))
}

func TestCreateSymmetricKey(t *testing.T) {
	length := 32

	result1, err1 := CreateSymmetricKey(length)
	result2, err2 := CreateSymmetricKey(length)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, result1, result2)
	assert.Equal(t, length, len(result1))
	assert.Equal(t, length, len(result2))
}
