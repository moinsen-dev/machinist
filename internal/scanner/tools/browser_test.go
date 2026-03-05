package tools

import (
	"context"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrowserScanner_Name(t *testing.T) {
	s := NewBrowserScanner(&util.MockCommandRunner{})
	assert.Equal(t, "browser", s.Name())
}

func TestBrowserScanner_Description(t *testing.T) {
	s := NewBrowserScanner(&util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestBrowserScanner_Category(t *testing.T) {
	s := NewBrowserScanner(&util.MockCommandRunner{})
	assert.Equal(t, "tools", s.Category())
}

func TestBrowserScanner_Scan_ChromeDefault(t *testing.T) {
	plistJSON := `{
		"LSHandlers": [
			{"LSHandlerURLScheme": "https", "LSHandlerRoleAll": "com.google.chrome"},
			{"LSHandlerURLScheme": "http", "LSHandlerRoleAll": "com.google.chrome"}
		]
	}`

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"plutil -convert json -o - $HOME/Library/Preferences/com.apple.LaunchServices/com.apple.launchservices.secure.plist": {
				Output: plistJSON,
			},
		},
	}

	s := NewBrowserScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Browser)
	assert.Equal(t, "Chrome", result.Browser.Default)
	assert.Equal(t, "Sign in to browser and sync extensions", result.Browser.ExtensionsChecklist)
}

func TestBrowserScanner_Scan_SafariDefault(t *testing.T) {
	plistJSON := `{
		"LSHandlers": [
			{"LSHandlerURLScheme": "https", "LSHandlerRoleAll": "com.apple.safari"}
		]
	}`

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"plutil -convert json -o - $HOME/Library/Preferences/com.apple.LaunchServices/com.apple.launchservices.secure.plist": {
				Output: plistJSON,
			},
		},
	}

	s := NewBrowserScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Browser)
	assert.Equal(t, "Safari", result.Browser.Default)
}

func TestBrowserScanner_Scan_PlistError(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewBrowserScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Browser)
	assert.Empty(t, result.Browser.Default)
	assert.Equal(t, "Sign in to browser and sync extensions", result.Browser.ExtensionsChecklist)
}

func TestBrowserScanner_Scan_UnknownBrowser(t *testing.T) {
	plistJSON := `{
		"LSHandlers": [
			{"LSHandlerURLScheme": "https", "LSHandlerRoleAll": "com.example.mybrowser"}
		]
	}`

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"plutil -convert json -o - $HOME/Library/Preferences/com.apple.LaunchServices/com.apple.launchservices.secure.plist": {
				Output: plistJSON,
			},
		},
	}

	s := NewBrowserScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Browser)
	// Unknown bundle ID returned as-is
	assert.Equal(t, "com.example.mybrowser", result.Browser.Default)
}

func TestParseDefaultBrowser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Chrome",
			input:    `{"LSHandlers": [{"LSHandlerURLScheme": "https", "LSHandlerRoleAll": "com.google.chrome"}]}`,
			expected: "Chrome",
		},
		{
			name:     "Firefox",
			input:    `{"LSHandlers": [{"LSHandlerURLScheme": "https", "LSHandlerRoleAll": "org.mozilla.firefox"}]}`,
			expected: "Firefox",
		},
		{
			name:     "Arc",
			input:    `{"LSHandlers": [{"LSHandlerURLScheme": "https", "LSHandlerRoleAll": "company.thebrowser.browser"}]}`,
			expected: "Arc",
		},
		{
			name:     "Brave",
			input:    `{"LSHandlers": [{"LSHandlerURLScheme": "https", "LSHandlerRoleAll": "com.brave.browser"}]}`,
			expected: "Brave",
		},
		{
			name:     "InvalidJSON",
			input:    `not json`,
			expected: "",
		},
		{
			name:     "NoHTTPS",
			input:    `{"LSHandlers": [{"LSHandlerURLScheme": "ftp", "LSHandlerRoleAll": "com.google.chrome"}]}`,
			expected: "",
		},
		{
			name:     "EmptyHandlers",
			input:    `{"LSHandlers": []}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDefaultBrowser(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
