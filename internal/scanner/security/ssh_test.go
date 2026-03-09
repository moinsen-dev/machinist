package security

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHScanner_Name(t *testing.T) {
	s := NewSSHScanner("/tmp")
	assert.Equal(t, "ssh", s.Name())
}

func TestSSHScanner_Description(t *testing.T) {
	s := NewSSHScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestSSHScanner_Category(t *testing.T) {
	s := NewSSHScanner("/tmp")
	assert.Equal(t, "security", s.Category())
}

func TestSSHScanner_Scan_NoSSHDir(t *testing.T) {
	// homeDir exists but has no .ssh subdirectory.
	homeDir := t.TempDir()

	s := NewSSHScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.SSH)
}

func TestSSHScanner_Scan_HappyPath(t *testing.T) {
	homeDir := t.TempDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0o700))

	// Create key pairs.
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_rsa"), []byte("PRIVATE KEY"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_rsa.pub"), []byte("PUBLIC KEY"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("PRIVATE KEY ED25519"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_ed25519.pub"), []byte("PUBLIC KEY ED25519"), 0o644))

	// Create config and known_hosts.
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "config"), []byte("Host *\n  ServerAliveInterval 60\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "known_hosts"), []byte("github.com ssh-rsa AAAAB3N..."), 0o600))

	s := NewSSHScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.SSH)

	ssh := result.SSH
	assert.True(t, ssh.Encrypted)
	assert.Equal(t, filepath.Join(".ssh", "config"), ssh.ConfigFile)
	assert.Equal(t, filepath.Join(".ssh", "known_hosts"), ssh.KnownHosts)

	// Only private keys (no .pub files).
	require.Len(t, ssh.Keys, 2)
	// Sort for deterministic comparison.
	sortedKeys := make([]string, len(ssh.Keys))
	copy(sortedKeys, ssh.Keys)
	sort.Strings(sortedKeys)
	assert.Equal(t, "id_ed25519", sortedKeys[0])
	assert.Equal(t, "id_rsa", sortedKeys[1])
}

func TestSSHScanner_Scan_NoConfigOrKnownHosts(t *testing.T) {
	homeDir := t.TempDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0o700))

	// Only a key file, no config or known_hosts.
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_ecdsa"), []byte("ECDSA KEY"), 0o600))

	s := NewSSHScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.SSH)

	assert.Empty(t, result.SSH.ConfigFile)
	assert.Empty(t, result.SSH.KnownHosts)
	require.Len(t, result.SSH.Keys, 1)
	assert.Equal(t, "id_ecdsa", result.SSH.Keys[0])
}

func TestSSHScanner_Scan_NoPubFilesInKeys(t *testing.T) {
	homeDir := t.TempDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0o700))

	// Create only .pub files — these should be ignored.
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_rsa.pub"), []byte("PUBLIC"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_ed25519.pub"), []byte("PUBLIC"), 0o644))

	s := NewSSHScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.SSH)
	assert.Empty(t, result.SSH.Keys)
}

func TestSSHScanner_Scan_NonKeyFilesIgnored(t *testing.T) {
	homeDir := t.TempDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0o700))

	// Files that don't start with "id_" should not be included in Keys.
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "authorized_keys"), []byte("keys"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_rsa"), []byte("PRIVATE"), 0o600))

	s := NewSSHScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.SSH)
	require.Len(t, result.SSH.Keys, 1)
	assert.Equal(t, "id_rsa", result.SSH.Keys[0])
}

func TestSSHScanner_Scan_EncryptedAlwaysTrue(t *testing.T) {
	homeDir := t.TempDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0o700))

	s := NewSSHScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.SSH)
	assert.True(t, result.SSH.Encrypted, "SSH section should always be marked encrypted")
}
