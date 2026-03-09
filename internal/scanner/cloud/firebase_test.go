package cloud

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirebaseScanner_Name(t *testing.T) {
	s := NewFirebaseScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "firebase", s.Name())
}

func TestFirebaseScanner_Description(t *testing.T) {
	s := NewFirebaseScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestFirebaseScanner_Category(t *testing.T) {
	s := NewFirebaseScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "cloud", s.Category())
}

func TestFirebaseScanner_Scan_NotInstalled(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}
	s := NewFirebaseScanner("/tmp", mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.Firebase)
}

func TestFirebaseScanner_Scan_PrimaryConfigDir(t *testing.T) {
	homeDir := t.TempDir()

	// Create primary config directory: ~/.config/firebase/
	firebaseDir := filepath.Join(homeDir, ".config", "firebase")
	require.NoError(t, os.MkdirAll(firebaseDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"firebase": {Output: "", Err: nil},
		},
	}

	s := NewFirebaseScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Firebase)
	assert.Equal(t, filepath.Join(".config", "firebase"), result.Firebase.ConfigDir)
}

func TestFirebaseScanner_Scan_FallbackConfigstore(t *testing.T) {
	homeDir := t.TempDir()

	// Create fallback configstore file: ~/.config/configstore/firebase-tools.json
	configstoreDir := filepath.Join(homeDir, ".config", "configstore")
	require.NoError(t, os.MkdirAll(configstoreDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configstoreDir, "firebase-tools.json"), []byte("{}"), 0o644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"firebase": {Output: "", Err: nil},
		},
	}

	s := NewFirebaseScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Firebase)
	assert.Equal(t, filepath.Join(".config", "configstore"), result.Firebase.ConfigDir)
}

func TestFirebaseScanner_Scan_InstalledNoConfig(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"firebase": {Output: "", Err: nil},
		},
	}

	s := NewFirebaseScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.Firebase)
}
