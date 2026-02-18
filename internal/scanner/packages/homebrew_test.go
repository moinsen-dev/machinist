package packages

import (
	"context"
	"fmt"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMock(responses map[string]util.MockResponse) *util.MockCommandRunner {
	return &util.MockCommandRunner{Responses: responses}
}

func TestHomebrewScanner_Name(t *testing.T) {
	s := NewHomebrewScanner(&util.MockCommandRunner{})
	assert.Equal(t, "homebrew", s.Name())
}

func TestHomebrewScanner_Description(t *testing.T) {
	s := NewHomebrewScanner(&util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestHomebrewScanner_Category(t *testing.T) {
	s := NewHomebrewScanner(&util.MockCommandRunner{})
	assert.Equal(t, "packages", s.Category())
}

func TestHomebrewScanner_Scan_Formulae(t *testing.T) {
	mock := newMock(map[string]util.MockResponse{
		"brew": {Output: "", Err: nil}, // IsInstalled check
		"brew list --formula --versions": {
			Output: "git 2.43.0\ncurl 8.5.0\njq 1.7.1",
		},
		"brew list --cask":     {Output: ""},
		"brew tap":             {Output: ""},
		"brew services list":   {Output: ""},
	})

	s := NewHomebrewScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Homebrew)
	require.Len(t, result.Homebrew.Formulae, 3)

	assert.Equal(t, "git", result.Homebrew.Formulae[0].Name)
	assert.Equal(t, "2.43.0", result.Homebrew.Formulae[0].Version)
	assert.Equal(t, "curl", result.Homebrew.Formulae[1].Name)
	assert.Equal(t, "8.5.0", result.Homebrew.Formulae[1].Version)
	assert.Equal(t, "jq", result.Homebrew.Formulae[2].Name)
	assert.Equal(t, "1.7.1", result.Homebrew.Formulae[2].Version)
}

func TestHomebrewScanner_Scan_Casks(t *testing.T) {
	mock := newMock(map[string]util.MockResponse{
		"brew":                          {Output: "", Err: nil},
		"brew list --formula --versions": {Output: ""},
		"brew list --cask": {
			Output: "firefox\nvisual-studio-code\niterm2",
		},
		"brew tap":           {Output: ""},
		"brew services list": {Output: ""},
	})

	s := NewHomebrewScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Homebrew)
	require.Len(t, result.Homebrew.Casks, 3)

	assert.Equal(t, "firefox", result.Homebrew.Casks[0].Name)
	assert.Empty(t, result.Homebrew.Casks[0].Version)
	assert.Equal(t, "visual-studio-code", result.Homebrew.Casks[1].Name)
	assert.Equal(t, "iterm2", result.Homebrew.Casks[2].Name)
}

func TestHomebrewScanner_Scan_Taps(t *testing.T) {
	mock := newMock(map[string]util.MockResponse{
		"brew":                          {Output: "", Err: nil},
		"brew list --formula --versions": {Output: ""},
		"brew list --cask":              {Output: ""},
		"brew tap": {
			Output: "homebrew/core\nhomebrew/cask\nhomebrew/services",
		},
		"brew services list": {Output: ""},
	})

	s := NewHomebrewScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Homebrew)
	require.Len(t, result.Homebrew.Taps, 3)

	assert.Equal(t, "homebrew/core", result.Homebrew.Taps[0])
	assert.Equal(t, "homebrew/cask", result.Homebrew.Taps[1])
	assert.Equal(t, "homebrew/services", result.Homebrew.Taps[2])
}

func TestHomebrewScanner_Scan_Services(t *testing.T) {
	mock := newMock(map[string]util.MockResponse{
		"brew":                          {Output: "", Err: nil},
		"brew list --formula --versions": {Output: ""},
		"brew list --cask":              {Output: ""},
		"brew tap":                      {Output: ""},
		"brew services list": {
			Output: "Name       Status  User    File\npostgresql started doedel ~/Library/LaunchAgents/homebrew.mxcl.postgresql.plist\nredis      started doedel ~/Library/LaunchAgents/homebrew.mxcl.redis.plist\nnginx      none",
		},
	})

	s := NewHomebrewScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Homebrew)
	require.Len(t, result.Homebrew.Services, 3)

	assert.Equal(t, "postgresql", result.Homebrew.Services[0].Name)
	assert.Equal(t, "started", result.Homebrew.Services[0].Status)
	assert.Equal(t, "redis", result.Homebrew.Services[1].Name)
	assert.Equal(t, "started", result.Homebrew.Services[1].Status)
	assert.Equal(t, "nginx", result.Homebrew.Services[2].Name)
	assert.Equal(t, "none", result.Homebrew.Services[2].Status)
}

func TestHomebrewScanner_Scan_Combined(t *testing.T) {
	mock := newMock(map[string]util.MockResponse{
		"brew": {Output: "", Err: nil},
		"brew list --formula --versions": {
			Output: "git 2.43.0\ncurl 8.5.0",
		},
		"brew list --cask": {
			Output: "firefox\niterm2",
		},
		"brew tap": {
			Output: "homebrew/core\nhomebrew/cask",
		},
		"brew services list": {
			Output: "Name       Status  User    File\npostgresql started doedel ~/Library/LaunchAgents/homebrew.mxcl.postgresql.plist",
		},
	})

	s := NewHomebrewScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Homebrew)

	assert.Len(t, result.Homebrew.Formulae, 2)
	assert.Len(t, result.Homebrew.Casks, 2)
	assert.Len(t, result.Homebrew.Taps, 2)
	assert.Len(t, result.Homebrew.Services, 1)

	assert.Equal(t, "homebrew", result.ScannerName)
}

func TestHomebrewScanner_Scan_BrewNotInstalled(t *testing.T) {
	// No "brew" key in Responses → IsInstalled returns false
	mock := newMock(map[string]util.MockResponse{})

	s := NewHomebrewScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Homebrew)
	assert.Equal(t, "homebrew", result.ScannerName)
}

func TestHomebrewScanner_Scan_FormulaCommandError(t *testing.T) {
	mock := newMock(map[string]util.MockResponse{
		"brew": {Output: "", Err: nil},
		"brew list --formula --versions": {
			Output: "",
			Err:    fmt.Errorf("brew formula list failed"),
		},
		"brew list --cask": {
			Output: "firefox\niterm2",
		},
		"brew tap": {
			Output: "homebrew/core",
		},
		"brew services list": {Output: ""},
	})

	s := NewHomebrewScanner(mock)
	result, err := s.Scan(context.Background())

	// The implementation returns nil + error when any brew sub-command fails.
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "brew formula list failed")
}

func TestHomebrewScanner_Scan_SingleFormula(t *testing.T) {
	mock := newMock(map[string]util.MockResponse{
		"brew": {Output: "", Err: nil},
		"brew list --formula --versions": {
			Output: "git 2.43.0",
		},
		"brew list --cask":   {Output: ""},
		"brew tap":           {Output: ""},
		"brew services list": {Output: ""},
	})

	s := NewHomebrewScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Homebrew)
	require.Len(t, result.Homebrew.Formulae, 1)
	assert.Equal(t, "git", result.Homebrew.Formulae[0].Name)
	assert.Equal(t, "2.43.0", result.Homebrew.Formulae[0].Version)
}

func TestHomebrewScanner_Scan_EmptyOutput(t *testing.T) {
	mock := newMock(map[string]util.MockResponse{
		"brew":                          {Output: "", Err: nil},
		"brew list --formula --versions": {Output: ""},
		"brew list --cask":              {Output: ""},
		"brew tap":                      {Output: ""},
		"brew services list":            {Output: ""},
	})

	s := NewHomebrewScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Homebrew)
	assert.Empty(t, result.Homebrew.Formulae)
	assert.Empty(t, result.Homebrew.Casks)
	assert.Empty(t, result.Homebrew.Taps)
	assert.Empty(t, result.Homebrew.Services)
}
