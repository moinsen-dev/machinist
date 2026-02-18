package editors

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

// ---------------------------------------------------------------------------
// VSCodeScanner tests
// ---------------------------------------------------------------------------

func TestVSCodeScanner_Name(t *testing.T) {
	s := NewVSCodeScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "vscode", s.Name())
}

func TestVSCodeScanner_Description(t *testing.T) {
	s := NewVSCodeScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestVSCodeScanner_Category(t *testing.T) {
	s := NewVSCodeScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "editors", s.Category())
}

func TestVSCodeScanner_Scan_Extensions(t *testing.T) {
	homeDir := t.TempDir()

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"code --list-extensions": {Output: "ms-python.python\ndbaeumer.vscode-eslint\nesbenp.prettier-vscode"},
		},
	}
	s := NewVSCodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.VSCode)
	assert.Len(t, result.VSCode.Extensions, 3)
	assert.Contains(t, result.VSCode.Extensions, "ms-python.python")
	assert.Contains(t, result.VSCode.Extensions, "dbaeumer.vscode-eslint")
	assert.Contains(t, result.VSCode.Extensions, "esbenp.prettier-vscode")
}

func TestVSCodeScanner_Scan_ConfigFiles(t *testing.T) {
	homeDir := t.TempDir()

	// Create VSCode config directory structure.
	configDir := filepath.Join(homeDir, "Library", "Application Support", "Code", "User")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	settingsContent := `{"editor.fontSize": 14}`
	keybindingsContent := `[]`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "settings.json"), []byte(settingsContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "keybindings.json"), []byte(keybindingsContent), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"code --list-extensions": {Output: "", Err: fmt.Errorf("not installed")},
		},
	}
	s := NewVSCodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.VSCode)
	assert.Len(t, result.VSCode.ConfigFiles, 2)

	// Verify paths and content hashes.
	sources := make(map[string]string)
	for _, cf := range result.VSCode.ConfigFiles {
		sources[cf.Source] = cf.ContentHash
	}
	assert.Contains(t, sources, "Library/Application Support/Code/User/settings.json")
	assert.Contains(t, sources, "Library/Application Support/Code/User/keybindings.json")

	for _, cf := range result.VSCode.ConfigFiles {
		assert.NotEmpty(t, cf.ContentHash, "ContentHash should not be empty for %s", cf.Source)
	}
}

func TestVSCodeScanner_Scan_SnippetsDir(t *testing.T) {
	homeDir := t.TempDir()

	// Create snippets directory.
	snippetsDir := filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "snippets")
	require.NoError(t, os.MkdirAll(snippetsDir, 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"code --list-extensions": {Output: "", Err: fmt.Errorf("not installed")},
		},
	}
	s := NewVSCodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.VSCode)
	assert.NotEmpty(t, result.VSCode.SnippetsDir)
	assert.Equal(t, "Library/Application Support/Code/User/snippets", result.VSCode.SnippetsDir)
}

func TestVSCodeScanner_Scan_Combined(t *testing.T) {
	homeDir := t.TempDir()

	// Create config files.
	configDir := filepath.Join(homeDir, "Library", "Application Support", "Code", "User")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "settings.json"), []byte(`{"editor.fontSize": 14}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "keybindings.json"), []byte(`[]`), 0644))

	// Create snippets directory.
	snippetsDir := filepath.Join(configDir, "snippets")
	require.NoError(t, os.MkdirAll(snippetsDir, 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"code --list-extensions": {Output: "ms-python.python\ndbaeumer.vscode-eslint\nesbenp.prettier-vscode"},
		},
	}
	s := NewVSCodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.VSCode)

	// Extensions populated.
	assert.Len(t, result.VSCode.Extensions, 3)

	// Config files populated.
	assert.Len(t, result.VSCode.ConfigFiles, 2)

	// Snippets dir populated.
	assert.NotEmpty(t, result.VSCode.SnippetsDir)
}

func TestVSCodeScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()

	// No config dirs, command fails.
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"code --list-extensions": {Output: "", Err: fmt.Errorf("not installed")},
		},
	}
	s := NewVSCodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	// VSCode section should be nil when nothing is detected.
	assert.Nil(t, result.VSCode)
}

func TestVSCodeScanner_Scan_NoExtensions(t *testing.T) {
	homeDir := t.TempDir()

	// Create config files so scanner detects VSCode.
	configDir := filepath.Join(homeDir, "Library", "Application Support", "Code", "User")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "settings.json"), []byte(`{}`), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"code --list-extensions": {Output: ""},
		},
	}
	s := NewVSCodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.VSCode)

	// Extensions should be empty.
	assert.Empty(t, result.VSCode.Extensions)

	// Config files should still be detected.
	assert.Len(t, result.VSCode.ConfigFiles, 1)
}

// ---------------------------------------------------------------------------
// CursorScanner tests
// ---------------------------------------------------------------------------

func TestCursorScanner_Name(t *testing.T) {
	s := NewCursorScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "cursor", s.Name())
}

func TestCursorScanner_Category(t *testing.T) {
	s := NewCursorScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "editors", s.Category())
}

func TestCursorScanner_Scan_Extensions(t *testing.T) {
	homeDir := t.TempDir()

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"cursor --list-extensions": {Output: "ms-python.python\ncontinue.continue"},
		},
	}
	s := NewCursorScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Cursor)
	assert.Len(t, result.Cursor.Extensions, 2)
	assert.Contains(t, result.Cursor.Extensions, "ms-python.python")
	assert.Contains(t, result.Cursor.Extensions, "continue.continue")
}

func TestCursorScanner_Scan_ConfigFiles(t *testing.T) {
	homeDir := t.TempDir()

	// Create Cursor config directory structure.
	configDir := filepath.Join(homeDir, "Library", "Application Support", "Cursor", "User")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "settings.json"), []byte(`{"editor.fontSize": 14}`), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"cursor --list-extensions": {Output: "", Err: fmt.Errorf("not installed")},
		},
	}
	s := NewCursorScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Cursor)
	assert.Len(t, result.Cursor.ConfigFiles, 1)
	assert.Equal(t, "Library/Application Support/Cursor/User/settings.json", result.Cursor.ConfigFiles[0].Source)
	assert.NotEmpty(t, result.Cursor.ConfigFiles[0].ContentHash)
}

func TestCursorScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"cursor --list-extensions": {Output: "", Err: fmt.Errorf("not installed")},
		},
	}
	s := NewCursorScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	// Cursor section should be nil when nothing is detected.
	assert.Nil(t, result.Cursor)
}
