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

func TestGCPScanner_Name(t *testing.T) {
	s := NewGCPScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "gcp", s.Name())
}

func TestGCPScanner_Description(t *testing.T) {
	s := NewGCPScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestGCPScanner_Category(t *testing.T) {
	s := NewGCPScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "cloud", s.Category())
}

func TestGCPScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{Responses: map[string]util.MockResponse{}}
	s := NewGCPScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.GCP)
}

func TestGCPScanner_Scan_InstalledNoConfig(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gcloud": {Output: "", Err: nil},
		},
	}
	s := NewGCPScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.GCP)
	assert.Empty(t, result.GCP.ConfigDir)
}

func TestGCPScanner_Scan_WithConfigDir(t *testing.T) {
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config", "gcloud")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gcloud": {Output: "", Err: nil},
		},
	}
	s := NewGCPScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GCP)
	assert.Equal(t, configDir, result.GCP.ConfigDir)
}
