package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"quasarflow-api/pkg/errors"
)

// Encryptor defines the interface for encryption operations
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AESEncryptor implements the Encryptor interface using AES encryption
type AESEncryptor struct {
	key []byte
}

func NewAESEncryptor(key string) (*AESEncryptor, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: got %d bytes", errors.ErrInvalidKeySize, len(key))
	}
	return &AESEncryptor{key: []byte(key)}, nil
}

func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errors.ErrEncryptionFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errors.ErrEncryptionFailed, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("%w: failed to generate nonce: %v", errors.ErrEncryptionFailed, err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("%w: invalid base64 encoding: %v", errors.ErrInvalidCiphertext, err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errors.ErrDecryptionFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errors.ErrDecryptionFailed, err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("%w: data too short", errors.ErrInvalidCiphertext)
	}

	nonce := data[:nonceSize]
	encryptedData := data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errors.ErrDecryptionFailed, err)
	}

	return string(plaintext), nil
}
