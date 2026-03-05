package cloud

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerScanner_Name(t *testing.T) {
	s := NewDockerScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "docker", s.Name())
}

func TestDockerScanner_Description(t *testing.T) {
	s := NewDockerScanner("/tmp", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestDockerScanner_Category(t *testing.T) {
	s := NewDockerScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "cloud", s.Category())
}

func TestDockerScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "docker" key absent -> IsInstalled returns false
		},
	}

	s := NewDockerScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Docker)
}

func TestDockerScanner_Scan_Full(t *testing.T) {
	homeDir := t.TempDir()

	// Create ~/.docker/config.json
	dockerDir := filepath.Join(homeDir, ".docker")
	require.NoError(t, os.MkdirAll(dockerDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dockerDir, "config.json"), []byte(`{}`), 0644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"docker":        {Output: "", Err: nil}, // IsInstalled check
			"colima status": {Output: "colima is running", Err: nil},
			"docker images --format {{.Repository}}:{{.Tag}}": {
				Output: "nginx:latest\nredis:7.0\n<none>:<none>\npostgres:15",
			},
		},
	}

	s := NewDockerScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Docker)

	docker := result.Docker
	assert.Equal(t, ".docker/config.json", docker.ConfigFile)
	assert.Equal(t, "colima", docker.Runtime)
	assert.Equal(t, []string{"nginx:latest", "redis:7.0", "postgres:15"}, docker.FrequentlyUsedImages)
}

func TestDockerScanner_Scan_OrbstackRuntime(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"docker":        {Output: "", Err: nil},
			"colima status": {Output: "", Err: fmt.Errorf("not running")},
			"orbctl status": {Output: "running", Err: nil},
			"docker images --format {{.Repository}}:{{.Tag}}": {
				Output: "node:20",
			},
		},
	}

	s := NewDockerScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Docker)
	assert.Equal(t, "orbstack", result.Docker.Runtime)
}

func TestDockerScanner_Scan_PodmanRuntime(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"docker":          {Output: "", Err: nil},
			"colima status":   {Output: "", Err: fmt.Errorf("not running")},
			"orbctl status":   {Output: "", Err: fmt.Errorf("not found")},
			"podman --version": {Output: "podman version 4.5.0", Err: nil},
			"docker images --format {{.Repository}}:{{.Tag}}": {
				Output: "",
			},
		},
	}

	s := NewDockerScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Docker)
	assert.Equal(t, "podman", result.Docker.Runtime)
}

func TestDockerScanner_Scan_DockerDesktopRuntime(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"docker":          {Output: "", Err: nil},
			"colima status":   {Output: "", Err: fmt.Errorf("not running")},
			"orbctl status":   {Output: "", Err: fmt.Errorf("not found")},
			"podman --version": {Output: "", Err: fmt.Errorf("not found")},
			"docker images --format {{.Repository}}:{{.Tag}}": {
				Output: "",
			},
		},
	}

	s := NewDockerScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Docker)
	assert.Equal(t, "docker-desktop", result.Docker.Runtime)
}

func TestDockerScanner_Scan_NoConfigFile(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"docker":          {Output: "", Err: nil},
			"colima status":   {Output: "", Err: fmt.Errorf("not running")},
			"orbctl status":   {Output: "", Err: fmt.Errorf("not found")},
			"podman --version": {Output: "", Err: fmt.Errorf("not found")},
			"docker images --format {{.Repository}}:{{.Tag}}": {
				Output: "alpine:3.18",
			},
		},
	}

	s := NewDockerScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Docker)

	assert.Empty(t, result.Docker.ConfigFile)
	assert.Equal(t, []string{"alpine:3.18"}, result.Docker.FrequentlyUsedImages)
}

func TestDockerScanner_Scan_ImagesCommandError(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"docker":          {Output: "", Err: nil},
			"colima status":   {Output: "running", Err: nil},
			"docker images --format {{.Repository}}:{{.Tag}}": {
				Output: "",
				Err:    fmt.Errorf("docker daemon not running"),
			},
		},
	}

	s := NewDockerScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Docker)

	// Images error is handled gracefully
	assert.Nil(t, result.Docker.FrequentlyUsedImages)
	assert.Equal(t, "colima", result.Docker.Runtime)
}
