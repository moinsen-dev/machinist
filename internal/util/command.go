package util

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CommandRunner abstracts shell command execution for testability.
type CommandRunner interface {
	// Run executes a command and returns its stdout trimmed.
	Run(ctx context.Context, name string, args ...string) (string, error)
	// RunLines executes a command and returns stdout split by newlines, empty lines removed.
	RunLines(ctx context.Context, name string, args ...string) ([]string, error)
	// IsInstalled checks if a command is available in PATH.
	IsInstalled(ctx context.Context, name string) bool
}

// RealCommandRunner uses os/exec for real command execution.
type RealCommandRunner struct{}

func (r *RealCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *RealCommandRunner) RunLines(ctx context.Context, name string, args ...string) ([]string, error) {
	output, err := r.Run(ctx, name, args...)
	if err != nil {
		return nil, err
	}
	return splitLines(output), nil
}

func (r *RealCommandRunner) IsInstalled(ctx context.Context, name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// MockResponse holds a predefined response for MockCommandRunner.
type MockResponse struct {
	Output string
	Err    error
}

// MockCommandRunner maps "name arg1 arg2" keys to predefined responses. For testing.
type MockCommandRunner struct {
	Responses map[string]MockResponse
	Calls     []string // records all commands called, for verification
}

func (m *MockCommandRunner) key(name string, args ...string) string {
	return strings.Join(append([]string{name}, args...), " ")
}

func (m *MockCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	k := m.key(name, args...)
	m.Calls = append(m.Calls, k)

	resp, ok := m.Responses[k]
	if !ok {
		return "", fmt.Errorf("mock: unknown command %q", k)
	}
	return resp.Output, resp.Err
}

func (m *MockCommandRunner) RunLines(ctx context.Context, name string, args ...string) ([]string, error) {
	output, err := m.Run(ctx, name, args...)
	if err != nil {
		return nil, err
	}
	return splitLines(output), nil
}

func (m *MockCommandRunner) IsInstalled(ctx context.Context, name string) bool {
	resp, ok := m.Responses[name]
	if !ok {
		return false
	}
	return resp.Err == nil
}

// splitLines splits a string by newlines and removes empty lines.
func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, "\n")
	var result []string
	for _, line := range parts {
		if line != "" {
			result = append(result, line)
		}
	}
	if result == nil {
		return []string{}
	}
	return result
}
