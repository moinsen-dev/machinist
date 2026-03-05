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

const simctlJSON = `{
  "devices": {
    "com.apple.CoreSimulator.SimRuntime.iOS-17-2": [
      {"name": "iPhone 15", "udid": "abc", "state": "Shutdown"},
      {"name": "iPhone 15 Pro", "udid": "def", "state": "Booted"},
      {"name": "iPad Pro (12.9-inch) (6th generation)", "udid": "ghi", "state": "Shutdown"}
    ],
    "com.apple.CoreSimulator.SimRuntime.iOS-16-4": [
      {"name": "iPhone 14", "udid": "jkl", "state": "Shutdown"},
      {"name": "iPhone 15", "udid": "mno", "state": "Shutdown"}
    ]
  }
}`

func TestXcodeScanner_Name(t *testing.T) {
	s := NewXcodeScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "xcode", s.Name())
}

func TestXcodeScanner_Description(t *testing.T) {
	s := NewXcodeScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestXcodeScanner_Category(t *testing.T) {
	s := NewXcodeScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "editors", s.Category())
}

func TestXcodeScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	// "xcode-select" key absent → IsInstalled returns false.
	cmd := &util.MockCommandRunner{Responses: map[string]util.MockResponse{}}
	s := NewXcodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Xcode)
}

func TestXcodeScanner_Scan_Simulators(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"xcode-select":                            {Output: "", Err: nil}, // IsInstalled
			"xcrun simctl list devices available -j": {Output: simctlJSON},
		},
	}
	s := NewXcodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Xcode)

	// Unique device names: iPhone 15, iPhone 15 Pro, iPad Pro..., iPhone 14.
	assert.Len(t, result.Xcode.Simulators, 4)
	assert.Contains(t, result.Xcode.Simulators, "iPhone 15")
	assert.Contains(t, result.Xcode.Simulators, "iPhone 15 Pro")
	assert.Contains(t, result.Xcode.Simulators, "iPad Pro (12.9-inch) (6th generation)")
	assert.Contains(t, result.Xcode.Simulators, "iPhone 14")
}

func TestXcodeScanner_Scan_SimulatorsDeduplication(t *testing.T) {
	homeDir := t.TempDir()
	// "iPhone 15" appears in two runtimes — should only be listed once.
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"xcode-select":                            {Output: "", Err: nil},
			"xcrun simctl list devices available -j": {Output: simctlJSON},
		},
	}
	s := NewXcodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Xcode)

	count := 0
	for _, name := range result.Xcode.Simulators {
		if name == "iPhone 15" {
			count++
		}
	}
	assert.Equal(t, 1, count, "iPhone 15 should appear exactly once despite appearing in two runtimes")
}

func TestXcodeScanner_Scan_SimctlError(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"xcode-select":                            {Output: "", Err: nil},
			"xcrun simctl list devices available -j": {Output: "", Err: fmt.Errorf("xcrun failed")},
		},
	}
	s := NewXcodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Xcode)
	// Simulators should be empty when xcrun fails.
	assert.Empty(t, result.Xcode.Simulators)
}

func TestXcodeScanner_Scan_ConfigFilePlist(t *testing.T) {
	homeDir := t.TempDir()
	prefsDir := filepath.Join(homeDir, "Library", "Preferences")
	require.NoError(t, os.MkdirAll(prefsDir, 0755))
	plistPath := filepath.Join(prefsDir, "com.apple.dt.Xcode.plist")
	require.NoError(t, os.WriteFile(plistPath, []byte("fake plist"), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"xcode-select":                            {Output: "", Err: nil},
			"xcrun simctl list devices available -j": {Output: `{"devices":{}}`},
		},
	}
	s := NewXcodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Xcode)
	require.Len(t, result.Xcode.ConfigFiles, 1)
	assert.Equal(t, "Library/Preferences/com.apple.dt.Xcode.plist", result.Xcode.ConfigFiles[0].Source)
	assert.Equal(t, "configs/xcode/com.apple.dt.Xcode.plist", result.Xcode.ConfigFiles[0].BundlePath)
}

func TestXcodeScanner_Scan_NoPlist(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"xcode-select":                            {Output: "", Err: nil},
			"xcrun simctl list devices available -j": {Output: `{"devices":{}}`},
		},
	}
	s := NewXcodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Xcode)
	assert.Empty(t, result.Xcode.ConfigFiles)
}

func TestXcodeScanner_Scan_EmptyDevices(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"xcode-select":                            {Output: "", Err: nil},
			"xcrun simctl list devices available -j": {Output: `{"devices":{}}`},
		},
	}
	s := NewXcodeScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Xcode)
	assert.Empty(t, result.Xcode.Simulators)
}
