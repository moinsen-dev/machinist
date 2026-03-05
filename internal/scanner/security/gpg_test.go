package security

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGPGScanner_Name(t *testing.T) {
	s := NewGPGScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "gpg", s.Name())
}

func TestGPGScanner_Description(t *testing.T) {
	s := NewGPGScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestGPGScanner_Category(t *testing.T) {
	s := NewGPGScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "security", s.Category())
}

func TestGPGScanner_Scan_NotInstalled(t *testing.T) {
	// No "gpg" key → IsInstalled returns false.
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}
	s := NewGPGScanner("/tmp", mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.GPG)
}

func TestGPGScanner_Scan_HappyPath(t *testing.T) {
	homeDir := t.TempDir()
	gnupgDir := filepath.Join(homeDir, ".gnupg")
	require.NoError(t, os.MkdirAll(gnupgDir, 0o700))

	// Create config files.
	require.NoError(t, os.WriteFile(filepath.Join(gnupgDir, "gpg.conf"), []byte("# gpg config\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(gnupgDir, "gpg-agent.conf"), []byte("# agent config\n"), 0o600))

	gpgOutput := "" +
		"pub:u:4096:1:ABCDEF1234567890:1680000000:::u:::scESC:\n" +
		"fpr:::::::::FINGERPRINT1234567890ABCDEF1234567890:\n" +
		"uid:u::::1680000000::HASH::Test User <test@example.com>:::::::::0:\n" +
		"pub:u:256:22:FEDCBA0987654321:1690000000:::u:::scESC:\n" +
		"fpr:::::::::FEDCBA0987654321ABCDEF1234567890FEDCBA:\n"

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gpg": {Output: "", Err: nil}, // IsInstalled
			"gpg --list-keys --keyid-format long --with-colons": {
				Output: gpgOutput,
			},
		},
	}

	s := NewGPGScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GPG)

	gpg := result.GPG
	assert.True(t, gpg.Encrypted)
	require.Len(t, gpg.Keys, 2)
	assert.Equal(t, "ABCDEF1234567890", gpg.Keys[0])
	assert.Equal(t, "FEDCBA0987654321", gpg.Keys[1])

	require.Len(t, gpg.ConfigFiles, 2)
	assert.Equal(t, ".gnupg/gpg.conf", gpg.ConfigFiles[0].Source)
	assert.Equal(t, "configs/gnupg/gpg.conf", gpg.ConfigFiles[0].BundlePath)
	assert.Equal(t, ".gnupg/gpg-agent.conf", gpg.ConfigFiles[1].Source)
	assert.Equal(t, "configs/gnupg/gpg-agent.conf", gpg.ConfigFiles[1].BundlePath)
}

func TestGPGScanner_Scan_NoConfigFiles(t *testing.T) {
	homeDir := t.TempDir()
	// ~/.gnupg does not exist — no config files recorded.

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gpg": {Output: "", Err: nil},
			"gpg --list-keys --keyid-format long --with-colons": {
				Output: "pub:u:4096:1:AABBCCDD11223344:1680000000:::u:::scESC:\n",
			},
		},
	}

	s := NewGPGScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GPG)
	require.Len(t, result.GPG.Keys, 1)
	assert.Equal(t, "AABBCCDD11223344", result.GPG.Keys[0])
	assert.Empty(t, result.GPG.ConfigFiles)
}

func TestGPGScanner_Scan_GPGListKeysError(t *testing.T) {
	homeDir := t.TempDir()
	gnupgDir := filepath.Join(homeDir, ".gnupg")
	require.NoError(t, os.MkdirAll(gnupgDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(gnupgDir, "gpg.conf"), []byte(""), 0o600))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gpg": {Output: "", Err: nil},
			"gpg --list-keys --keyid-format long --with-colons": {
				Output: "", Err: fmt.Errorf("gpg: error reading key"),
			},
		},
	}

	s := NewGPGScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GPG)
	// No keys parsed but config file still detected.
	assert.Empty(t, result.GPG.Keys)
	require.Len(t, result.GPG.ConfigFiles, 1)
	assert.Equal(t, ".gnupg/gpg.conf", result.GPG.ConfigFiles[0].Source)
}

func TestGPGScanner_Scan_NonPubLinesIgnored(t *testing.T) {
	homeDir := t.TempDir()

	// Output with uid, sub, fpr lines — only pub lines should produce keys.
	gpgOutput := "" +
		"tru::1:1680000000:0:3:1:5\n" +
		"pub:u:4096:1:KEYID00000000001:1680000000:::u:::scESC:\n" +
		"uid:u::::1680000000::HASH::Alice <alice@example.com>:::::::::0:\n" +
		"sub:u:4096:1:SUBKEYID1234567:1680000000::::::e::::::23:\n"

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gpg": {Output: "", Err: nil},
			"gpg --list-keys --keyid-format long --with-colons": {
				Output: gpgOutput,
			},
		},
	}

	s := NewGPGScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GPG)
	// Only one pub line → one key.
	require.Len(t, result.GPG.Keys, 1)
	assert.Equal(t, "KEYID00000000001", result.GPG.Keys[0])
}

func TestGPGScanner_Scan_EncryptedAlwaysTrue(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gpg": {Output: "", Err: nil},
			"gpg --list-keys --keyid-format long --with-colons": {
				Output: "",
			},
		},
	}

	s := NewGPGScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GPG)
	assert.True(t, result.GPG.Encrypted, "GPG section should always be marked encrypted")
}
