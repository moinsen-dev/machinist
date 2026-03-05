package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvFilesScanner_Name(t *testing.T) {
	s := NewEnvFilesScanner("/tmp")
	assert.Equal(t, "env-files", s.Name())
}

func TestEnvFilesScanner_Description(t *testing.T) {
	s := NewEnvFilesScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestEnvFilesScanner_Category(t *testing.T) {
	s := NewEnvFilesScanner("/tmp")
	assert.Equal(t, "tools", s.Category())
}

func TestEnvFilesScanner_Scan_NoWorkspaceDirs(t *testing.T) {
	homeDir := t.TempDir()
	s := NewEnvFilesScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.EnvFiles)
}

func TestEnvFilesScanner_Scan_FindsEnvFiles(t *testing.T) {
	homeDir := t.TempDir()

	// Create workspace with projects containing .env files
	projectDir := filepath.Join(homeDir, "Code", "myproject")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, ".env"), []byte("KEY=val"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, ".env.local"), []byte("LOCAL=val"), 0o644))

	s := NewEnvFilesScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.EnvFiles)
	assert.True(t, result.EnvFiles.Encrypted)
	assert.Len(t, result.EnvFiles.Files, 2)
}

func TestEnvFilesScanner_Scan_SkipsHiddenDirs(t *testing.T) {
	homeDir := t.TempDir()

	// Create a hidden directory with .env file inside
	hiddenDir := filepath.Join(homeDir, "Code", ".hidden-project")
	require.NoError(t, os.MkdirAll(hiddenDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(hiddenDir, ".env"), []byte("SECRET=val"), 0o644))

	// Create a visible project with .env file
	visibleDir := filepath.Join(homeDir, "Code", "visible-project")
	require.NoError(t, os.MkdirAll(visibleDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(visibleDir, ".env"), []byte("KEY=val"), 0o644))

	s := NewEnvFilesScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.EnvFiles)
	assert.Len(t, result.EnvFiles.Files, 1)
	assert.Contains(t, result.EnvFiles.Files[0].Source, "visible-project")
}

func TestEnvFilesScanner_Scan_RespectsDepthLimit(t *testing.T) {
	homeDir := t.TempDir()

	// Level 1 project (within depth)
	level1 := filepath.Join(homeDir, "Code", "proj1")
	require.NoError(t, os.MkdirAll(level1, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(level1, ".env"), []byte("L1=val"), 0o644))

	// Level 2 project (within depth)
	level2 := filepath.Join(homeDir, "Code", "org", "proj2")
	require.NoError(t, os.MkdirAll(level2, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(level2, ".env"), []byte("L2=val"), 0o644))

	// Level 3 project (beyond depth, should be skipped)
	level3 := filepath.Join(homeDir, "Code", "org", "sub", "proj3")
	require.NoError(t, os.MkdirAll(level3, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(level3, ".env"), []byte("L3=val"), 0o644))

	s := NewEnvFilesScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.EnvFiles)
	// Should find level 1 and level 2 but not level 3
	assert.Len(t, result.EnvFiles.Files, 2)
}

func TestEnvFilesScanner_Scan_MultipleWorkspaceDirs(t *testing.T) {
	homeDir := t.TempDir()

	// Create .env files in multiple workspace directories
	for _, dir := range []string{"Code", "Projects"} {
		projectDir := filepath.Join(homeDir, dir, "myapp")
		require.NoError(t, os.MkdirAll(projectDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, ".env"), []byte("KEY=val"), 0o644))
	}

	s := NewEnvFilesScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.EnvFiles)
	assert.Len(t, result.EnvFiles.Files, 2)
}

func TestEnvFilesScanner_Scan_EnvProduction(t *testing.T) {
	homeDir := t.TempDir()

	projectDir := filepath.Join(homeDir, "work", "app")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, ".env.production"), []byte("PROD=true"), 0o644))

	s := NewEnvFilesScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.EnvFiles)
	require.Len(t, result.EnvFiles.Files, 1)
	assert.Contains(t, result.EnvFiles.Files[0].Source, ".env.production")
}
