package runtimes

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJavaScanner_Name(t *testing.T) {
	s := NewJavaScanner("/home/user", &util.MockCommandRunner{})
	assert.Equal(t, "java", s.Name())
}

func TestJavaScanner_Description(t *testing.T) {
	s := NewJavaScanner("/home/user", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestJavaScanner_Category(t *testing.T) {
	s := NewJavaScanner("/home/user", &util.MockCommandRunner{})
	assert.Equal(t, "runtimes", s.Category())
}

func TestJavaScanner_Scan_WithSDKMAN(t *testing.T) {
	homeDir := t.TempDir()

	// Build SDKMAN directory structure.
	javaCandidatesDir := filepath.Join(homeDir, ".sdkman", "candidates", "java")
	require.NoError(t, os.MkdirAll(javaCandidatesDir, 0o755))

	// Create two installed Java version directories.
	for _, ver := range []string{"21.0.1-tem", "17.0.9-tem"} {
		require.NoError(t, os.MkdirAll(filepath.Join(javaCandidatesDir, ver), 0o755))
	}

	// Create a "current" symlink pointing to the default version.
	defaultVerDir := filepath.Join(javaCandidatesDir, "21.0.1-tem")
	currentLink := filepath.Join(javaCandidatesDir, "current")
	require.NoError(t, os.Symlink(defaultVerDir, currentLink))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewJavaScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Java)

	java := result.Java
	assert.Equal(t, "sdkman", java.Manager)
	assert.Contains(t, java.Versions, "21.0.1-tem")
	assert.Contains(t, java.Versions, "17.0.9-tem")
	// "current" symlink should not appear in the versions list.
	assert.NotContains(t, java.Versions, "current")
	assert.Equal(t, "21.0.1-tem", java.DefaultVersion)
	assert.NotEmpty(t, java.JavaHome)
}

func TestJavaScanner_Scan_SystemJava(t *testing.T) {
	homeDir := t.TempDir()
	// No .sdkman directory.

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"java": {Output: "", Err: nil}, // IsInstalled check
			"java -version": {
				Output: `openjdk version "21.0.1" 2023-10-17` + "\n" +
					`OpenJDK Runtime Environment (build 21.0.1+12-29)`,
			},
		},
	}

	s := NewJavaScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Java)

	java := result.Java
	assert.Empty(t, java.Manager)
	assert.Equal(t, []string{"21.0.1"}, java.Versions)
	assert.Equal(t, "21.0.1", java.DefaultVersion)
}

func TestJavaScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	// No .sdkman directory, no java in PATH.

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "java" key absent → IsInstalled returns false
		},
	}

	s := NewJavaScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Java)
}

func TestJavaScanner_Scan_SDKMANNoVersions(t *testing.T) {
	homeDir := t.TempDir()

	// .sdkman exists but no Java candidates installed.
	javaCandidatesDir := filepath.Join(homeDir, ".sdkman", "candidates", "java")
	require.NoError(t, os.MkdirAll(javaCandidatesDir, 0o755))

	// No java in PATH either.
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewJavaScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	// Should fall back to system java, which is also not available.
	assert.Nil(t, result.Java)
}

func TestParseJavaVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`openjdk version "21.0.1" 2023-10-17`, "21.0.1"},
		{`java version "1.8.0_392"`, "1.8.0_392"},
		{`openjdk version "17.0.9" 2023-10-17` + "\nOpenJDK Runtime Environment", "17.0.9"},
		{"no version here", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseJavaVersion(tt.input))
		})
	}
}
