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

func TestFlyioScanner_Name(t *testing.T) {
	s := NewFlyioScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "flyio", s.Name())
}

func TestFlyioScanner_Description(t *testing.T) {
	s := NewFlyioScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestFlyioScanner_Category(t *testing.T) {
	s := NewFlyioScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "cloud", s.Category())
}

func TestFlyioScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{Responses: map[string]util.MockResponse{}}
	s := NewFlyioScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Flyio)
}

func TestFlyioScanner_Scan_InstalledNoConfig(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"fly": {Output: "", Err: nil},
		},
	}
	s := NewFlyioScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Flyio)
	assert.Empty(t, result.Flyio.ConfigFile)
}

func TestFlyioScanner_Scan_WithConfigFile(t *testing.T) {
	homeDir := t.TempDir()
	flyDir := filepath.Join(homeDir, ".fly")
	require.NoError(t, os.MkdirAll(flyDir, 0755))
	configFile := filepath.Join(flyDir, "config.yml")
	require.NoError(t, os.WriteFile(configFile, []byte("access_token: test"), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"fly": {Output: "", Err: nil},
		},
	}
	s := NewFlyioScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Flyio)
	assert.Equal(t, configFile, result.Flyio.ConfigFile)
}
