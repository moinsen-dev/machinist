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

func TestKubernetesScanner_Name(t *testing.T) {
	s := NewKubernetesScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "kubernetes", s.Name())
}

func TestKubernetesScanner_Description(t *testing.T) {
	s := NewKubernetesScanner("/tmp", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestKubernetesScanner_Category(t *testing.T) {
	s := NewKubernetesScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "cloud", s.Category())
}

func TestKubernetesScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "kubectl" key absent -> IsInstalled returns false
		},
	}

	s := NewKubernetesScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Kubernetes)
}

func TestKubernetesScanner_Scan_Full(t *testing.T) {
	homeDir := t.TempDir()

	// Create ~/.kube/config
	kubeDir := filepath.Join(homeDir, ".kube")
	require.NoError(t, os.MkdirAll(kubeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(kubeDir, "config"), []byte("apiVersion: v1"), 0644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"kubectl":                                   {Output: "", Err: nil}, // IsInstalled check
			"kubectl config get-contexts -o name": {Output: "minikube\nproduction\nstaging"},
		},
	}

	s := NewKubernetesScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Kubernetes)

	k8s := result.Kubernetes
	assert.Equal(t, ".kube/config", k8s.ConfigFile)
	assert.Equal(t, []string{"minikube", "production", "staging"}, k8s.Contexts)
}

func TestKubernetesScanner_Scan_NoConfigFile(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"kubectl":                                   {Output: "", Err: nil},
			"kubectl config get-contexts -o name": {Output: "docker-desktop"},
		},
	}

	s := NewKubernetesScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Kubernetes)

	assert.Empty(t, result.Kubernetes.ConfigFile)
	assert.Equal(t, []string{"docker-desktop"}, result.Kubernetes.Contexts)
}

func TestKubernetesScanner_Scan_ContextsError(t *testing.T) {
	homeDir := t.TempDir()

	// Create config file
	kubeDir := filepath.Join(homeDir, ".kube")
	require.NoError(t, os.MkdirAll(kubeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(kubeDir, "config"), []byte("apiVersion: v1"), 0644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"kubectl":                                   {Output: "", Err: nil},
			"kubectl config get-contexts -o name": {Output: "", Err: fmt.Errorf("error getting contexts")},
		},
	}

	s := NewKubernetesScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Kubernetes)

	assert.Equal(t, ".kube/config", result.Kubernetes.ConfigFile)
	// Contexts error handled gracefully
	assert.Nil(t, result.Kubernetes.Contexts)
}
