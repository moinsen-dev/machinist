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

func TestAPIToolsScanner_Name(t *testing.T) {
	s := NewAPIToolsScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "api-tools", s.Name())
}

func TestAPIToolsScanner_Description(t *testing.T) {
	s := NewAPIToolsScanner("/tmp", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestAPIToolsScanner_Category(t *testing.T) {
	s := NewAPIToolsScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "tools", s.Category())
}

func TestAPIToolsScanner_Scan_NothingFound(t *testing.T) {
	homeDir := t.TempDir()
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewAPIToolsScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.APITools)
}

func TestAPIToolsScanner_Scan_PostmanOnly(t *testing.T) {
	homeDir := t.TempDir()
	postmanDir := filepath.Join(homeDir, "Library", "Application Support", "Postman")
	require.NoError(t, os.MkdirAll(postmanDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewAPIToolsScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.APITools)
	require.Len(t, result.APITools.ConfigFiles, 1)
	assert.Equal(t, "Postman", result.APITools.ConfigFiles[0].Source)
	assert.Equal(t, "configs/Postman", result.APITools.ConfigFiles[0].BundlePath)
	assert.False(t, result.APITools.Mkcert)
}

func TestAPIToolsScanner_Scan_InsomniaOnly(t *testing.T) {
	homeDir := t.TempDir()
	insomniaDir := filepath.Join(homeDir, "Library", "Application Support", "Insomnia")
	require.NoError(t, os.MkdirAll(insomniaDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewAPIToolsScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.APITools)
	require.Len(t, result.APITools.ConfigFiles, 1)
	assert.Equal(t, "Insomnia", result.APITools.ConfigFiles[0].Source)
	assert.Equal(t, "configs/Insomnia", result.APITools.ConfigFiles[0].BundlePath)
}

func TestAPIToolsScanner_Scan_MkcertOnly(t *testing.T) {
	homeDir := t.TempDir()
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"mkcert": {Output: "", Err: nil},
		},
	}

	s := NewAPIToolsScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.APITools)
	assert.True(t, result.APITools.Mkcert)
	assert.Empty(t, result.APITools.ConfigFiles)
}

func TestAPIToolsScanner_Scan_AllFound(t *testing.T) {
	homeDir := t.TempDir()
	postmanDir := filepath.Join(homeDir, "Library", "Application Support", "Postman")
	insomniaDir := filepath.Join(homeDir, "Library", "Application Support", "Insomnia")
	require.NoError(t, os.MkdirAll(postmanDir, 0o755))
	require.NoError(t, os.MkdirAll(insomniaDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"mkcert": {Output: "", Err: nil},
		},
	}

	s := NewAPIToolsScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.APITools)
	assert.Len(t, result.APITools.ConfigFiles, 2)
	assert.True(t, result.APITools.Mkcert)
}
