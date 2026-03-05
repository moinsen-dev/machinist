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

func TestAzureScanner_Name(t *testing.T) {
	s := NewAzureScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "azure", s.Name())
}

func TestAzureScanner_Description(t *testing.T) {
	s := NewAzureScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestAzureScanner_Category(t *testing.T) {
	s := NewAzureScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "cloud", s.Category())
}

func TestAzureScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{Responses: map[string]util.MockResponse{}}
	s := NewAzureScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Azure)
}

func TestAzureScanner_Scan_InstalledNoConfig(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"az": {Output: "", Err: nil},
		},
	}
	s := NewAzureScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Azure)
	assert.Empty(t, result.Azure.ConfigDir)
}

func TestAzureScanner_Scan_WithConfigDir(t *testing.T) {
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".azure")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"az": {Output: "", Err: nil},
		},
	}
	s := NewAzureScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Azure)
	assert.Equal(t, configDir, result.Azure.ConfigDir)
}
