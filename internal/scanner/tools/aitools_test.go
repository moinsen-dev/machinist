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

func TestAIToolsScanner_Name(t *testing.T) {
	s := NewAIToolsScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "ai-tools", s.Name())
}

func TestAIToolsScanner_Description(t *testing.T) {
	s := NewAIToolsScanner("/tmp", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestAIToolsScanner_Category(t *testing.T) {
	s := NewAIToolsScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "tools", s.Category())
}

func TestAIToolsScanner_Scan_BothFound(t *testing.T) {
	homeDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"ollama":      {Output: "", Err: nil},
			"ollama list": {Output: "NAME           ID            SIZE    MODIFIED\nllama3:latest  abc123        4.7 GB  2 days ago\nmistral:7b     def456        4.1 GB  5 days ago"},
		},
	}

	s := NewAIToolsScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.AITools)
	assert.Equal(t, filepath.Join(homeDir, ".claude"), result.AITools.ClaudeCodeConfig)
	assert.Equal(t, []string{"llama3:latest", "mistral:7b"}, result.AITools.OllamaModels)
}

func TestAIToolsScanner_Scan_OnlyClaude(t *testing.T) {
	homeDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewAIToolsScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.AITools)
	assert.Equal(t, filepath.Join(homeDir, ".claude"), result.AITools.ClaudeCodeConfig)
	assert.Empty(t, result.AITools.OllamaModels)
}

func TestAIToolsScanner_Scan_OnlyOllama(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"ollama":      {Output: "", Err: nil},
			"ollama list": {Output: "NAME           ID            SIZE    MODIFIED\ncodellama:7b   ghi789        3.8 GB  1 day ago"},
		},
	}

	s := NewAIToolsScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.AITools)
	assert.Empty(t, result.AITools.ClaudeCodeConfig)
	assert.Equal(t, []string{"codellama:7b"}, result.AITools.OllamaModels)
}

func TestAIToolsScanner_Scan_NothingFound(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewAIToolsScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.AITools)
}

func TestAIToolsScanner_Scan_OllamaNoModels(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"ollama":      {Output: "", Err: nil},
			"ollama list": {Output: "NAME           ID            SIZE    MODIFIED"},
		},
	}

	s := NewAIToolsScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	// No claude dir, no ollama models — should be nil
	assert.Nil(t, result.AITools)
}
