package system

import (
	"context"
	"fmt"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginItemsScanner_Name(t *testing.T) {
	s := NewLoginItemsScanner(&util.MockCommandRunner{})
	assert.Equal(t, "login-items", s.Name())
}

func TestLoginItemsScanner_Description(t *testing.T) {
	s := NewLoginItemsScanner(&util.MockCommandRunner{})
	assert.Equal(t, "Scans macOS login items", s.Description())
}

func TestLoginItemsScanner_Category(t *testing.T) {
	s := NewLoginItemsScanner(&util.MockCommandRunner{})
	assert.Equal(t, "system", s.Category())
}

func TestLoginItemsScanner_Scan_HappyPath(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			`osascript -e tell application "System Events" to get name of every login item`: {
				Output: "Dropbox, Alfred 5, Docker",
			},
		},
	}

	s := NewLoginItemsScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.LoginItems)
	assert.Equal(t, []string{"Dropbox", "Alfred 5", "Docker"}, result.LoginItems.Apps)
}

func TestLoginItemsScanner_Scan_SingleItem(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			`osascript -e tell application "System Events" to get name of every login item`: {
				Output: "Docker",
			},
		},
	}

	s := NewLoginItemsScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.LoginItems)
	assert.Equal(t, []string{"Docker"}, result.LoginItems.Apps)
}

func TestLoginItemsScanner_Scan_EmptyOutput(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			`osascript -e tell application "System Events" to get name of every login item`: {
				Output: "",
			},
		},
	}

	s := NewLoginItemsScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.LoginItems)
	assert.Nil(t, result.LoginItems.Apps)
}

func TestLoginItemsScanner_Scan_TCCError(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			`osascript -e tell application "System Events" to get name of every login item`: {
				Err: fmt.Errorf("execution error: Not authorized to send Apple events"),
			},
		},
	}

	s := NewLoginItemsScanner(mock)
	result, err := s.Scan(context.Background())

	// Should not error, but return empty section
	require.NoError(t, err)
	require.NotNil(t, result.LoginItems)
	assert.Nil(t, result.LoginItems.Apps)
}

func TestParseLoginItems(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"multiple", "Dropbox, Alfred 5, Docker", []string{"Dropbox", "Alfred 5", "Docker"}},
		{"single", "Docker", []string{"Docker"}},
		{"with spaces", "  Dropbox ,  Alfred 5 ", []string{"Dropbox", "Alfred 5"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseLoginItems(tt.input))
		})
	}
}
