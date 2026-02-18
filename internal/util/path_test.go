package util

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "tilde prefix is expanded to home dir",
			path: "~/foo",
			want: filepath.Join(home, "foo"),
		},
		{
			name: "tilde alone expands to home dir",
			path: "~",
			want: home,
		},
		{
			name: "absolute path returned unchanged",
			path: "/absolute/path",
			want: "/absolute/path",
		},
		{
			name: "relative path returned unchanged",
			path: "relative/path",
			want: "relative/path",
		},
		{
			name: "empty string returned unchanged",
			path: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandHome(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "fileexists_test_*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing file returns true",
			path: tmpFile.Name(),
			want: true,
		},
		{
			name: "nonexistent path returns false",
			path: "/nonexistent/path/to/file.txt",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FileExists(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDirExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "direxists_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing directory returns true",
			path: tmpDir,
			want: true,
		},
		{
			name: "nonexistent directory returns false",
			path: "/nonexistent/dir",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DirExists(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContentHash(t *testing.T) {
	// Create a temp file with known content
	tmpFile, err := os.CreateTemp("", "contenthash_test_*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte("hello world")
	_, err = tmpFile.Write(content)
	require.NoError(t, err)
	tmpFile.Close()

	expectedHash := fmt.Sprintf("%x", sha256.Sum256(content))

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "returns consistent SHA256 hex for known content",
			path:    tmpFile.Name(),
			want:    expectedHash,
			wantErr: false,
		},
		{
			name:    "returns error for nonexistent file",
			path:    "/nonexistent/file.txt",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ContentHash(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}

	// Verify consistency: calling twice gives the same result
	t.Run("consistent across calls", func(t *testing.T) {
		hash1, err := ContentHash(tmpFile.Name())
		require.NoError(t, err)
		hash2, err := ContentHash(tmpFile.Name())
		require.NoError(t, err)
		assert.Equal(t, hash1, hash2)
	})
}

func TestContentHash_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "contenthash_empty_*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	got, err := ContentHash(tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", got)
}

func TestDirExists_OnFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "direxists_file_*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	assert.False(t, DirExists(tmpFile.Name()))
}

func TestFileExists_OnDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fileexists_dir_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	assert.False(t, FileExists(tmpDir))
}
