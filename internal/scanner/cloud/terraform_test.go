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

func TestTerraformScanner_Name(t *testing.T) {
	s := NewTerraformScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "terraform", s.Name())
}

func TestTerraformScanner_Description(t *testing.T) {
	s := NewTerraformScanner("/tmp", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestTerraformScanner_Category(t *testing.T) {
	s := NewTerraformScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "cloud", s.Category())
}

func TestTerraformScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "terraform" key absent -> IsInstalled returns false
		},
	}

	s := NewTerraformScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Terraform)
}

func TestTerraformScanner_Scan_WithTerraformrc(t *testing.T) {
	homeDir := t.TempDir()

	// Create ~/.terraformrc
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".terraformrc"), []byte("plugin_cache_dir = \"$HOME/.terraform.d/plugin-cache\""), 0644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"terraform": {Output: "", Err: nil}, // IsInstalled check
		},
	}

	s := NewTerraformScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Terraform)

	assert.Equal(t, ".terraformrc", result.Terraform.ConfigFile)
}

func TestTerraformScanner_Scan_WithTerraformDir(t *testing.T) {
	homeDir := t.TempDir()

	// Create ~/.terraform.d/ directory (no .terraformrc)
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".terraform.d"), 0755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"terraform": {Output: "", Err: nil},
		},
	}

	s := NewTerraformScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Terraform)

	assert.Equal(t, ".terraform.d/", result.Terraform.ConfigFile)
}

func TestTerraformScanner_Scan_TerraformrcTakesPrecedence(t *testing.T) {
	homeDir := t.TempDir()

	// Create both ~/.terraformrc and ~/.terraform.d/
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".terraformrc"), []byte("{}"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".terraform.d"), 0755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"terraform": {Output: "", Err: nil},
		},
	}

	s := NewTerraformScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Terraform)

	// .terraformrc should take precedence over .terraform.d/
	assert.Equal(t, ".terraformrc", result.Terraform.ConfigFile)
}

func TestTerraformScanner_Scan_NoConfig(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"terraform": {Output: "", Err: nil},
		},
	}

	s := NewTerraformScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Terraform)

	// No config found, but section is still populated (terraform is installed)
	assert.Empty(t, result.Terraform.ConfigFile)
}
