package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptor(t *testing.T) {
	// Generate a test key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	// Create encryptor
	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)
	require.NotNil(t, encryptor)

	t.Run("Invalid Key Size", func(t *testing.T) {
		_, err := NewEncryptor([]byte("too short"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "encryption key must be 32 bytes")
	})

	t.Run("Encrypt and Decrypt", func(t *testing.T) {
		testCases := []struct {
			name      string
			plaintext string
		}{
			{
				name:      "Empty String",
				plaintext: "",
			},
			{
				name:      "Short String",
				plaintext: "hello",
			},
			{
				name:      "Long String",
				plaintext: "this is a very long string that needs to be encrypted",
			},
			{
				name:      "Special Characters",
				plaintext: "!@#$%^&*()_+-=[]{}|;:,.<>?",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Encrypt
				ciphertext, err := encryptor.Encrypt(tc.plaintext)
				require.NoError(t, err)
				require.NotEmpty(t, ciphertext)

				// Decrypt
				decrypted, err := encryptor.Decrypt(ciphertext)
				require.NoError(t, err)
				assert.Equal(t, tc.plaintext, decrypted)
			})
		}
	})

	t.Run("Invalid Ciphertext", func(t *testing.T) {
		_, err := encryptor.Decrypt("invalid base64")
		assert.Error(t, err)

		_, err = encryptor.Decrypt("aW52YWxpZA==") // valid base64 but invalid ciphertext
		assert.Error(t, err)
	})
}
