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

func TestCloudflareScanner_Name(t *testing.T) {
	s := NewCloudflareScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "cloudflare", s.Name())
}

func TestCloudflareScanner_Description(t *testing.T) {
	s := NewCloudflareScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestCloudflareScanner_Category(t *testing.T) {
	s := NewCloudflareScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "cloud", s.Category())
}

func TestCloudflareScanner_Scan_NotInstalled(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}
	s := NewCloudflareScanner("/tmp", mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.CloudflareWrangler)
}

func TestCloudflareScanner_Scan_PrimaryConfigDir(t *testing.T) {
	homeDir := t.TempDir()

	// Create primary config directory: ~/.config/.wrangler/
	wranglerDir := filepath.Join(homeDir, ".config", ".wrangler")
	require.NoError(t, os.MkdirAll(wranglerDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"wrangler": {Output: "", Err: nil},
		},
	}

	s := NewCloudflareScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.CloudflareWrangler)
	assert.Equal(t, filepath.Join(".config", ".wrangler"), result.CloudflareWrangler.ConfigDir)
}

func TestCloudflareScanner_Scan_FallbackWranglerConfig(t *testing.T) {
	homeDir := t.TempDir()

	// Create fallback directory: ~/.wrangler/config/
	fallbackDir := filepath.Join(homeDir, ".wrangler", "config")
	require.NoError(t, os.MkdirAll(fallbackDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"wrangler": {Output: "", Err: nil},
		},
	}

	s := NewCloudflareScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.CloudflareWrangler)
	assert.Equal(t, filepath.Join(".wrangler", "config"), result.CloudflareWrangler.ConfigDir)
}

func TestCloudflareScanner_Scan_InstalledNoConfig(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"wrangler": {Output: "", Err: nil},
		},
	}

	s := NewCloudflareScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.CloudflareWrangler)
}
