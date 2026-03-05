package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistriesScanner_Name(t *testing.T) {
	s := NewRegistriesScanner("/tmp")
	assert.Equal(t, "registries", s.Name())
}

func TestRegistriesScanner_Description(t *testing.T) {
	s := NewRegistriesScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestRegistriesScanner_Category(t *testing.T) {
	s := NewRegistriesScanner("/tmp")
	assert.Equal(t, "tools", s.Category())
}

func TestRegistriesScanner_Scan_AllFound(t *testing.T) {
	homeDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".npmrc"), []byte("data"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".pip"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".pip", "pip.conf"), []byte("data"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".cargo"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".cargo", "config.toml"), []byte("data"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".gemrc"), []byte("data"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".cocoapods"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".cocoapods", "config.yaml"), []byte("data"), 0o600))

	s := NewRegistriesScanner(homeDir)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Registries)
	assert.Len(t, result.Registries.ConfigFiles, 5)

	// .npmrc should be sensitive
	assert.True(t, result.Registries.ConfigFiles[0].Sensitive)
	// Others should not be sensitive
	for _, cf := range result.Registries.ConfigFiles[1:] {
		assert.False(t, cf.Sensitive, "expected %s to not be sensitive", cf.Source)
	}
}

func TestRegistriesScanner_Scan_NoneFound(t *testing.T) {
	homeDir := t.TempDir()

	s := NewRegistriesScanner(homeDir)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Registries)
}

func TestRegistriesScanner_Scan_PartialFound(t *testing.T) {
	homeDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".gemrc"), []byte("data"), 0o600))

	s := NewRegistriesScanner(homeDir)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Registries)
	assert.Len(t, result.Registries.ConfigFiles, 1)
	assert.Equal(t, ".gemrc", result.Registries.ConfigFiles[0].Source)
	assert.False(t, result.Registries.ConfigFiles[0].Sensitive)
}
