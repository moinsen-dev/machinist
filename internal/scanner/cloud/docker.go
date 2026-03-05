package cloud

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// DockerScanner scans Docker configuration, runtime, and frequently used images.
type DockerScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewDockerScanner creates a new DockerScanner with the given home directory and CommandRunner.
func NewDockerScanner(homeDir string, cmd util.CommandRunner) *DockerScanner {
	return &DockerScanner{homeDir: homeDir, cmd: cmd}
}

func (s *DockerScanner) Name() string        { return "docker" }
func (s *DockerScanner) Description() string { return "Scans Docker configuration, runtime, and images" }
func (s *DockerScanner) Category() string    { return "cloud" }

// Scan checks for Docker installation, config file, runtime, and frequently used images.
func (s *DockerScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	if !s.cmd.IsInstalled(ctx, "docker") {
		return result, nil
	}

	section := &domain.DockerSection{}

	// Check for ~/.docker/config.json
	configPath := filepath.Join(s.homeDir, ".docker", "config.json")
	if util.FileExists(configPath) {
		section.ConfigFile = ".docker/config.json"
	}

	// Detect container runtime
	section.Runtime = s.detectRuntime(ctx)

	// Get frequently used images
	output, err := s.cmd.Run(ctx, "docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	if err == nil {
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.Contains(line, "<none>") {
				continue
			}
			section.FrequentlyUsedImages = append(section.FrequentlyUsedImages, line)
		}
	}

	result.Docker = section
	return result, nil
}

// detectRuntime tries to determine which container runtime is being used.
func (s *DockerScanner) detectRuntime(ctx context.Context) string {
	if _, err := s.cmd.Run(ctx, "colima", "status"); err == nil {
		return "colima"
	}
	if _, err := s.cmd.Run(ctx, "orbctl", "status"); err == nil {
		return "orbstack"
	}
	if _, err := s.cmd.Run(ctx, "podman", "--version"); err == nil {
		return "podman"
	}
	return "docker-desktop"
}
