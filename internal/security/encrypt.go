package security

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"filippo.io/age/armor"
)

const ageArmorHeader = "-----BEGIN AGE ENCRYPTED FILE-----"

// Encrypt encrypts data using age with a passphrase (scrypt recipient).
// The output is ASCII-armored for text-safe storage.
func Encrypt(data []byte, passphrase string) ([]byte, error) {
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return nil, fmt.Errorf("creating scrypt recipient: %w", err)
	}

	var buf bytes.Buffer
	armorWriter := armor.NewWriter(&buf)

	writer, err := age.Encrypt(armorWriter, recipient)
	if err != nil {
		return nil, fmt.Errorf("initializing encryption: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("writing encrypted data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("finalizing encryption: %w", err)
	}

	if err := armorWriter.Close(); err != nil {
		return nil, fmt.Errorf("finalizing armor: %w", err)
	}

	return buf.Bytes(), nil
}

// Decrypt decrypts age-encrypted data with a passphrase.
func Decrypt(data []byte, passphrase string) ([]byte, error) {
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, fmt.Errorf("creating scrypt identity: %w", err)
	}

	armorReader := armor.NewReader(bytes.NewReader(data))

	reader, err := age.Decrypt(armorReader, identity)
	if err != nil {
		return nil, fmt.Errorf("initializing decryption: %w", err)
	}

	decrypted, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading decrypted data: %w", err)
	}

	return decrypted, nil
}

// EncryptFile reads the source file, encrypts it, and writes to the destination.
func EncryptFile(src, dst, passphrase string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	encrypted, err := Encrypt(data, passphrase)
	if err != nil {
		return err
	}

	if err := os.WriteFile(dst, encrypted, 0o600); err != nil {
		return fmt.Errorf("writing encrypted file: %w", err)
	}

	return nil
}

// DecryptFile reads the encrypted source file, decrypts it, and writes to the destination.
func DecryptFile(src, dst, passphrase string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading encrypted file: %w", err)
	}

	decrypted, err := Decrypt(data, passphrase)
	if err != nil {
		return err
	}

	if err := os.WriteFile(dst, decrypted, 0o600); err != nil {
		return fmt.Errorf("writing decrypted file: %w", err)
	}

	return nil
}

// IsEncrypted checks if data starts with the age armor header.
func IsEncrypted(data []byte) bool {
	return bytes.HasPrefix(data, []byte(ageArmorHeader))
}
