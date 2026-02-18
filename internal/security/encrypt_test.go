package security

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncrypt_RoundTrip(t *testing.T) {
	plaintext := []byte("hello, machinist!")
	passphrase := "test-passphrase-123"

	encrypted, err := Encrypt(plaintext, passphrase)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := Decrypt(encrypted, passphrase)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncrypt_WrongPassphrase(t *testing.T) {
	plaintext := []byte("secret data")

	encrypted, err := Encrypt(plaintext, "correct")
	require.NoError(t, err)

	_, err = Decrypt(encrypted, "wrong")
	assert.Error(t, err)
}

func TestEncrypt_EmptyData(t *testing.T) {
	passphrase := "some-passphrase"

	encrypted, err := Encrypt([]byte{}, passphrase)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := Decrypt(encrypted, passphrase)
	require.NoError(t, err)
	assert.Equal(t, []byte{}, decrypted)
}

func TestEncryptFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "plain.txt")
	encPath := filepath.Join(dir, "encrypted.age")
	decPath := filepath.Join(dir, "decrypted.txt")

	content := []byte("file content for round-trip test")
	require.NoError(t, os.WriteFile(srcPath, content, 0o644))

	passphrase := "file-passphrase"

	err := EncryptFile(srcPath, encPath, passphrase)
	require.NoError(t, err)

	encData, err := os.ReadFile(encPath)
	require.NoError(t, err)
	assert.NotEqual(t, content, encData)

	err = DecryptFile(encPath, decPath, passphrase)
	require.NoError(t, err)

	decData, err := os.ReadFile(decPath)
	require.NoError(t, err)
	assert.Equal(t, content, decData)
}

func TestEncryptFile_SourceNotFound(t *testing.T) {
	dir := t.TempDir()
	err := EncryptFile(filepath.Join(dir, "nonexistent.txt"), filepath.Join(dir, "out.age"), "pass")
	assert.Error(t, err)
}

func TestDecryptFile_SourceNotFound(t *testing.T) {
	dir := t.TempDir()
	err := DecryptFile(filepath.Join(dir, "nonexistent.age"), filepath.Join(dir, "out.txt"), "pass")
	assert.Error(t, err)
}

func TestIsEncrypted(t *testing.T) {
	plaintext := []byte("just plain text")
	passphrase := "check-passphrase"

	encrypted, err := Encrypt(plaintext, passphrase)
	require.NoError(t, err)

	assert.True(t, IsEncrypted(encrypted), "encrypted data should be detected as encrypted")
	assert.False(t, IsEncrypted(plaintext), "plaintext should not be detected as encrypted")
	assert.False(t, IsEncrypted([]byte{}), "empty data should not be detected as encrypted")
}
