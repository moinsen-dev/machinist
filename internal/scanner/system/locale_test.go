package system

import (
	"context"
	"fmt"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleScanner_Name(t *testing.T) {
	s := NewLocaleScanner(&util.MockCommandRunner{})
	assert.Equal(t, "locale", s.Name())
}

func TestLocaleScanner_Description(t *testing.T) {
	s := NewLocaleScanner(&util.MockCommandRunner{})
	assert.Equal(t, "Scans macOS locale, timezone, and hostname settings", s.Description())
}

func TestLocaleScanner_Category(t *testing.T) {
	s := NewLocaleScanner(&util.MockCommandRunner{})
	assert.Equal(t, "system", s.Category())
}

func TestLocaleScanner_Scan_HappyPath(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"defaults read NSGlobalDomain AppleLanguages": {Output: "(\n    \"en-US\",\n    \"de-DE\"\n)"},
			"defaults read NSGlobalDomain AppleLocale":    {Output: "en_US"},
			"systemsetup -gettimezone":                    {Output: "Time Zone: America/Los_Angeles"},
			"scutil --get ComputerName":                   {Output: "My MacBook Pro"},
			"scutil --get LocalHostName":                  {Output: "My-MacBook-Pro"},
		},
	}

	s := NewLocaleScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Locale)

	assert.Equal(t, "en-US", result.Locale.Language)
	assert.Equal(t, "en_US", result.Locale.Region)
	assert.Equal(t, "America/Los_Angeles", result.Locale.Timezone)
	assert.Equal(t, "My MacBook Pro", result.Locale.ComputerName)
	assert.Equal(t, "My-MacBook-Pro", result.Locale.LocalHostname)
}

func TestLocaleScanner_Scan_AllErrors(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"defaults read NSGlobalDomain AppleLanguages": {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain AppleLocale":    {Err: fmt.Errorf("not set")},
			"systemsetup -gettimezone":                    {Err: fmt.Errorf("permission denied")},
			"scutil --get ComputerName":                   {Err: fmt.Errorf("not set")},
			"scutil --get LocalHostName":                  {Err: fmt.Errorf("not set")},
		},
	}

	s := NewLocaleScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Locale)
	// All fields should be empty
	assert.Empty(t, result.Locale.Language)
	assert.Empty(t, result.Locale.Region)
	assert.Empty(t, result.Locale.Timezone)
	assert.Empty(t, result.Locale.ComputerName)
	assert.Empty(t, result.Locale.LocalHostname)
}

func TestLocaleScanner_Scan_PartialData(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"defaults read NSGlobalDomain AppleLanguages": {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain AppleLocale":    {Output: "en_GB"},
			"systemsetup -gettimezone":                    {Output: "Time Zone: Europe/London"},
			"scutil --get ComputerName":                   {Err: fmt.Errorf("not set")},
			"scutil --get LocalHostName":                  {Output: "dev-machine"},
		},
	}

	s := NewLocaleScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Locale)
	assert.Empty(t, result.Locale.Language)
	assert.Equal(t, "en_GB", result.Locale.Region)
	assert.Equal(t, "Europe/London", result.Locale.Timezone)
	assert.Empty(t, result.Locale.ComputerName)
	assert.Equal(t, "dev-machine", result.Locale.LocalHostname)
}

func TestParseFirstLanguage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"standard", "(\n    \"en-US\",\n    \"de-DE\"\n)", "en-US"},
		{"single", "(\n    \"ja-JP\"\n)", "ja-JP"},
		{"empty", "()", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseFirstLanguage(tt.input))
		})
	}
}

func TestParseTimezone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"standard", "Time Zone: America/Los_Angeles", "America/Los_Angeles"},
		{"no prefix", "Europe/Berlin", "Europe/Berlin"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseTimezone(tt.input))
		})
	}
}
