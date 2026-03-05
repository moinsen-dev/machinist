package tools

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// AIToolsScanner scans AI developer tool configurations.
type AIToolsScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewAIToolsScanner creates a new AIToolsScanner.
func NewAIToolsScanner(homeDir string, cmd util.CommandRunner) *AIToolsScanner {
	return &AIToolsScanner{homeDir: homeDir, cmd: cmd}
}

func (s *AIToolsScanner) Name() string        { return "ai-tools" }
func (s *AIToolsScanner) Description() string  { return "Scans AI developer tool configurations" }
func (s *AIToolsScanner) Category() string     { return "tools" }

// Scan checks for Claude Code config and Ollama models.
func (s *AIToolsScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	section := &domain.AIToolsSection{}
	found := false

	// Check for Claude Code configuration
	claudeDir := filepath.Join(s.homeDir, ".claude")
	if util.DirExists(claudeDir) {
		section.ClaudeCodeConfig = claudeDir
		found = true
	}

	// Check for Ollama
	if s.cmd.IsInstalled(ctx, "ollama") {
		models, err := s.parseOllamaModels(ctx)
		if err == nil && len(models) > 0 {
			section.OllamaModels = models
			found = true
		}
	}

	if found {
		result.AITools = section
	}

	return result, nil
}

// parseOllamaModels runs `ollama list` and parses model names from the output.
func (s *AIToolsScanner) parseOllamaModels(ctx context.Context) ([]string, error) {
	output, err := s.cmd.Run(ctx, "ollama", "list")
	if err != nil {
		return nil, err
	}

	var models []string
	for i, line := range strings.Split(output, "\n") {
		// Skip the header line
		if i == 0 {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// First column is the model name
		fields := strings.Fields(line)
		if len(fields) > 0 {
			models = append(models, fields[0])
		}
	}

	return models, nil
}
