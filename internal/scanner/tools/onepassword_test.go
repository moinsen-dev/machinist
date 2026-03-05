package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnePasswordScanner_Name(t *testing.T) {
	s := NewOnePasswordScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "1password", s.Name())
}

func TestOnePasswordScanner_Description(t *testing.T) {
	s := NewOnePasswordScanner("/tmp", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestOnePasswordScanner_Category(t *testing.T) {
	s := NewOnePasswordScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "tools", s.Category())
}

func TestOnePasswordScanner_Scan_Found(t *testing.T) {
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config", "op")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"op": {Output: "", Err: nil},
		},
	}

	s := NewOnePasswordScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.OnePassword)
	assert.Equal(t, configDir, result.OnePassword.ConfigDir)
}

func TestOnePasswordScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewOnePasswordScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.OnePassword)
}

func TestOnePasswordScanner_Scan_NoConfigDir(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"op": {Output: "", Err: nil},
		},
	}

	s := NewOnePasswordScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.OnePassword)
}
