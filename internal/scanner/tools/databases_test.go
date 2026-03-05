package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabasesScanner_Name(t *testing.T) {
	s := NewDatabasesScanner("/tmp")
	assert.Equal(t, "databases", s.Name())
}

func TestDatabasesScanner_Description(t *testing.T) {
	s := NewDatabasesScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestDatabasesScanner_Category(t *testing.T) {
	s := NewDatabasesScanner("/tmp")
	assert.Equal(t, "tools", s.Category())
}

func TestDatabasesScanner_Scan_AllFound(t *testing.T) {
	homeDir := t.TempDir()

	// Create .pgpass
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".pgpass"), []byte("data"), 0o600))

	// Create .my.cnf
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".my.cnf"), []byte("data"), 0o600))

	// Create TablePlus dir
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, "Library", "Application Support", "com.tinyapp.TablePlus"), 0o755))

	// Create .dbeaver4 dir
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".dbeaver4"), 0o755))

	s := NewDatabasesScanner(homeDir)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Databases)
	assert.Len(t, result.Databases.ConfigFiles, 4)

	// Verify sensitive flags
	assert.True(t, result.Databases.ConfigFiles[0].Sensitive)  // .pgpass
	assert.True(t, result.Databases.ConfigFiles[1].Sensitive)  // .my.cnf
	assert.False(t, result.Databases.ConfigFiles[2].Sensitive) // TablePlus
	assert.False(t, result.Databases.ConfigFiles[3].Sensitive) // DBeaver
}

func TestDatabasesScanner_Scan_NoneFound(t *testing.T) {
	homeDir := t.TempDir()

	s := NewDatabasesScanner(homeDir)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Databases)
}

func TestDatabasesScanner_Scan_PartialFound(t *testing.T) {
	homeDir := t.TempDir()

	// Create only .pgpass
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".pgpass"), []byte("data"), 0o600))

	s := NewDatabasesScanner(homeDir)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Databases)
	assert.Len(t, result.Databases.ConfigFiles, 1)
	assert.Equal(t, ".pgpass", result.Databases.ConfigFiles[0].Source)
	assert.Equal(t, "configs/.pgpass", result.Databases.ConfigFiles[0].BundlePath)
	assert.True(t, result.Databases.ConfigFiles[0].Sensitive)
}
