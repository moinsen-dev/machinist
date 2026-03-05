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

func TestVercelScanner_Name(t *testing.T) {
	s := NewVercelScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "vercel", s.Name())
}

func TestVercelScanner_Description(t *testing.T) {
	s := NewVercelScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestVercelScanner_Category(t *testing.T) {
	s := NewVercelScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "cloud", s.Category())
}

func TestVercelScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{Responses: map[string]util.MockResponse{}}
	s := NewVercelScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Vercel)
}

func TestVercelScanner_Scan_InstalledNoConfig(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"vercel": {Output: "", Err: nil},
		},
	}
	s := NewVercelScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Vercel)
	assert.Empty(t, result.Vercel.ConfigDir)
}

func TestVercelScanner_Scan_WithConfigDir(t *testing.T) {
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".vercel")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"vercel": {Output: "", Err: nil},
		},
	}
	s := NewVercelScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Vercel)
	assert.Equal(t, configDir, result.Vercel.ConfigDir)
}
