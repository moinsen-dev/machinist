package system

import (
	"context"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppsScanner_Name(t *testing.T) {
	s := NewAppsScanner(&util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "apps", s.Name())
}

func TestAppsScanner_Category(t *testing.T) {
	s := NewAppsScanner(&util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "system", s.Category())
}

func TestAppsScanner_Scan_MasApps(t *testing.T) {
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"mas": {Output: "", Err: nil}, // IsInstalled checks this key
			"mas list": {
				Output: "497799835  Xcode (15.2)\n409183694  Keynote (14.0)\n1295203466  Microsoft Remote Desktop (10.9.5)",
			},
		},
	}

	s := NewAppsScanner(cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Apps)
	require.Len(t, result.Apps.AppStore, 3)

	// Verify first app: Xcode
	assert.Equal(t, 497799835, result.Apps.AppStore[0].ID)
	assert.Equal(t, "Xcode", result.Apps.AppStore[0].Name)
	assert.Equal(t, "mas", result.Apps.AppStore[0].Source)

	// Verify second app: Keynote
	assert.Equal(t, 409183694, result.Apps.AppStore[1].ID)
	assert.Equal(t, "Keynote", result.Apps.AppStore[1].Name)
	assert.Equal(t, "mas", result.Apps.AppStore[1].Source)

	// Verify third app: Microsoft Remote Desktop
	assert.Equal(t, 1295203466, result.Apps.AppStore[2].ID)
	assert.Equal(t, "Microsoft Remote Desktop", result.Apps.AppStore[2].Name)
	assert.Equal(t, "mas", result.Apps.AppStore[2].Source)
}

func TestAppsScanner_Scan_MasNotInstalled(t *testing.T) {
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "mas" key absent means IsInstalled returns false
		},
	}

	s := NewAppsScanner(cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	// Apps section should be nil when mas is not installed.
	assert.Nil(t, result.Apps)
}

func TestAppsScanner_Scan_NoApps(t *testing.T) {
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"mas":      {Output: "", Err: nil},
			"mas list": {Output: ""},
		},
	}

	s := NewAppsScanner(cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	// When mas is installed but returns no apps, Apps should be nil (nothing found).
	assert.Nil(t, result.Apps)
}

func TestAppsScanner_Scan_MasParsingEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		masOutput  string
		wantCount  int
		wantID     int
		wantName   string
	}{
		{
			name:      "single digit ID",
			masOutput: "1  TestApp (1.0)",
			wantCount: 1,
			wantID:    1,
			wantName:  "TestApp",
		},
		{
			name:      "app name with parentheses in name",
			masOutput: "123456  Some App (Pro) (2.1.0)",
			wantCount: 1,
			wantID:    123456,
			wantName:  "Some App (Pro)",
		},
		{
			name:      "extra whitespace between ID and name",
			masOutput: "999999999   Extra Spaces (1.0)",
			wantCount: 1,
			wantID:    999999999,
			wantName:  "Extra Spaces",
		},
		{
			name:      "trailing newlines and blank lines",
			masOutput: "\n497799835  Xcode (15.2)\n\n409183694  Keynote (14.0)\n\n",
			wantCount: 2,
			wantID:    497799835,
			wantName:  "Xcode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &util.MockCommandRunner{
				Responses: map[string]util.MockResponse{
					"mas":      {Output: "", Err: nil},
					"mas list": {Output: tt.masOutput},
				},
			}

			s := NewAppsScanner(cmd)
			result, err := s.Scan(context.Background())

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.wantCount == 0 {
				assert.Nil(t, result.Apps)
				return
			}

			require.NotNil(t, result.Apps)
			require.Len(t, result.Apps.AppStore, tt.wantCount)
			assert.Equal(t, tt.wantID, result.Apps.AppStore[0].ID)
			assert.Equal(t, tt.wantName, result.Apps.AppStore[0].Name)
			assert.Equal(t, "mas", result.Apps.AppStore[0].Source)
		})
	}
}
