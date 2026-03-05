package cloud

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// KubernetesScanner scans Kubernetes configuration and contexts.
type KubernetesScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewKubernetesScanner creates a new KubernetesScanner with the given home directory and CommandRunner.
func NewKubernetesScanner(homeDir string, cmd util.CommandRunner) *KubernetesScanner {
	return &KubernetesScanner{homeDir: homeDir, cmd: cmd}
}

func (s *KubernetesScanner) Name() string        { return "kubernetes" }
func (s *KubernetesScanner) Description() string { return "Scans Kubernetes configuration and contexts" }
func (s *KubernetesScanner) Category() string    { return "cloud" }

// Scan checks for kubectl installation, kubeconfig file, and configured contexts.
func (s *KubernetesScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	if !s.cmd.IsInstalled(ctx, "kubectl") {
		return result, nil
	}

	section := &domain.KubernetesSection{}

	// Check for ~/.kube/config
	configPath := filepath.Join(s.homeDir, ".kube", "config")
	if util.FileExists(configPath) {
		section.ConfigFile = ".kube/config"
	}

	// List contexts
	contexts, err := s.cmd.RunLines(ctx, "kubectl", "config", "get-contexts", "-o", "name")
	if err == nil {
		section.Contexts = contexts
	}

	result.Kubernetes = section
	return result, nil
}
