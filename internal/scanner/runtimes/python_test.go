package runtimes

import (
	"context"
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPythonScanner_Name(t *testing.T) {
	s := NewPythonScanner(&util.MockCommandRunner{})
	assert.Equal(t, "python", s.Name())
}

func TestPythonScanner_Description(t *testing.T) {
	s := NewPythonScanner(&util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestPythonScanner_Category(t *testing.T) {
	s := NewPythonScanner(&util.MockCommandRunner{})
	assert.Equal(t, "runtimes", s.Category())
}

func TestPythonScanner_Scan_WithPyenv(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"pyenv": {Output: "", Err: nil}, // IsInstalled returns true
			"pyenv versions --bare": {
				Output: "3.11.7\n3.12.1\n3.10.13",
			},
			"pyenv global": {
				Output: "3.12.1",
			},
			"pip list --format=json": {
				Output: `[{"name":"black","version":"24.1.0"},{"name":"ruff","version":"0.1.14"}]`,
			},
		},
	}

	s := NewPythonScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Python)

	py := result.Python
	assert.Equal(t, "pyenv", py.Manager)
	assert.Len(t, py.Versions, 3)
	assert.Contains(t, py.Versions, "3.11.7")
	assert.Contains(t, py.Versions, "3.12.1")
	assert.Contains(t, py.Versions, "3.10.13")
	assert.Equal(t, "3.12.1", py.DefaultVersion)
	assert.Len(t, py.GlobalPackages, 2)
	assert.Equal(t, "black", py.GlobalPackages[0].Name)
	assert.Equal(t, "24.1.0", py.GlobalPackages[0].Version)
	assert.Equal(t, "ruff", py.GlobalPackages[1].Name)
	assert.Equal(t, "0.1.14", py.GlobalPackages[1].Version)
}

func TestPythonScanner_Scan_WithUv(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// pyenv not installed (no entry means IsInstalled returns false)
			"uv": {Output: "", Err: nil}, // IsInstalled returns true
			"uv python list --only-installed": {
				Output: "cpython-3.12.1-macos-aarch64-none    /Users/user/.local/share/uv/python/cpython-3.12.1/bin/python3\ncpython-3.11.7-macos-aarch64-none    /Users/user/.local/share/uv/python/cpython-3.11.7/bin/python3",
			},
			"pip list --format=json": {
				Output: `[]`,
			},
		},
	}

	s := NewPythonScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Python)

	py := result.Python
	assert.Equal(t, "uv", py.Manager)
	assert.Len(t, py.Versions, 2)
	assert.Contains(t, py.Versions, "3.12.1")
	assert.Contains(t, py.Versions, "3.11.7")
}

func TestPythonScanner_Scan_SystemPython(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// pyenv and uv not installed (no entries)
			"python3 --version": {
				Output: "Python 3.12.1",
			},
			"pip list --format=json": {
				Output: `[]`,
			},
		},
	}

	s := NewPythonScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Python)

	py := result.Python
	assert.Equal(t, "", py.Manager)
	assert.Equal(t, "3.12.1", py.DefaultVersion)
}

func TestPythonScanner_Scan_NoPython(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// Nothing installed — all commands will fail via mock
		},
	}

	s := NewPythonScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.Python)
}

func TestPythonScanner_Scan_PipListParsing(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"pyenv": {Output: "", Err: nil},
			"pyenv versions --bare": {
				Output: "3.12.1",
			},
			"pyenv global": {
				Output: "3.12.1",
			},
			"pip list --format=json": {
				Output: `[
					{"name":"pip","version":"24.0"},
					{"name":"setuptools","version":"69.0.3"},
					{"name":"wheel","version":"0.42.0"},
					{"name":"black","version":"24.1.0"},
					{"name":"ruff","version":"0.1.14"}
				]`,
			},
		},
	}

	s := NewPythonScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Python)

	py := result.Python
	require.Len(t, py.GlobalPackages, 5)

	expected := []domain.Package{
		{Name: "pip", Version: "24.0"},
		{Name: "setuptools", Version: "69.0.3"},
		{Name: "wheel", Version: "0.42.0"},
		{Name: "black", Version: "24.1.0"},
		{Name: "ruff", Version: "0.1.14"},
	}

	for i, exp := range expected {
		assert.Equal(t, exp.Name, py.GlobalPackages[i].Name, "package %d name", i)
		assert.Equal(t, exp.Version, py.GlobalPackages[i].Version, "package %d version", i)
	}
}
