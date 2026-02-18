package system

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

func TestScheduledScanner_Name(t *testing.T) {
	s := NewScheduledScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "scheduled", s.Name())
}

func TestScheduledScanner_Category(t *testing.T) {
	s := NewScheduledScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "system", s.Category())
}

func TestScheduledScanner_Scan_Crontab(t *testing.T) {
	tmpDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"crontab -l": {
				Output: "# daily backup\n0 2 * * * /usr/local/bin/backup.sh\n*/15 * * * * /usr/local/bin/healthcheck.sh",
			},
		},
	}

	s := NewScheduledScanner(tmpDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Crontab)
	assert.Len(t, result.Crontab.Entries, 2)
	assert.Equal(t, "0 2 * * * /usr/local/bin/backup.sh", result.Crontab.Entries[0])
	assert.Equal(t, "*/15 * * * * /usr/local/bin/healthcheck.sh", result.Crontab.Entries[1])
}

func TestScheduledScanner_Scan_NoCrontab(t *testing.T) {
	tmpDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"crontab -l": {
				Output: "",
				Err:    fmt.Errorf("crontab: no crontab for user"),
			},
		},
	}

	s := NewScheduledScanner(tmpDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	// Crontab section should still exist but with empty entries
	if result.Crontab != nil {
		assert.Empty(t, result.Crontab.Entries)
	}
}

func TestScheduledScanner_Scan_LaunchAgents(t *testing.T) {
	tmpDir := t.TempDir()
	launchAgentsDir := filepath.Join(tmpDir, "Library", "LaunchAgents")
	require.NoError(t, os.MkdirAll(launchAgentsDir, 0o755))

	// Create sample plist files.
	require.NoError(t, os.WriteFile(
		filepath.Join(launchAgentsDir, "com.user.backup.plist"),
		[]byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict><key>Label</key><string>com.user.backup</string></dict></plist>`),
		0o644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(launchAgentsDir, "com.user.sync.plist"),
		[]byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict><key>Label</key><string>com.user.sync</string></dict></plist>`),
		0o644,
	))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"crontab -l": {
				Output: "",
				Err:    fmt.Errorf("crontab: no crontab for user"),
			},
		},
	}

	s := NewScheduledScanner(tmpDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.LaunchAgents)
	assert.Len(t, result.LaunchAgents.Plists, 2)

	// Collect names for order-independent assertion.
	names := make(map[string]bool)
	for _, p := range result.LaunchAgents.Plists {
		names[filepath.Base(p.Source)] = true
		assert.Contains(t, p.Source, "Library/LaunchAgents/")
	}
	assert.True(t, names["com.user.backup.plist"])
	assert.True(t, names["com.user.sync.plist"])
}

func TestScheduledScanner_Scan_NoLaunchAgents(t *testing.T) {
	tmpDir := t.TempDir()
	// No Library/LaunchAgents directory created.

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"crontab -l": {
				Output: "",
				Err:    fmt.Errorf("crontab: no crontab for user"),
			},
		},
	}

	s := NewScheduledScanner(tmpDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	if result.LaunchAgents != nil {
		assert.Empty(t, result.LaunchAgents.Plists)
	}
}

func TestScheduledScanner_Scan_Combined(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up LaunchAgents directory with one plist.
	launchAgentsDir := filepath.Join(tmpDir, "Library", "LaunchAgents")
	require.NoError(t, os.MkdirAll(launchAgentsDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(launchAgentsDir, "com.user.backup.plist"),
		[]byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict><key>Label</key><string>com.user.backup</string></dict></plist>`),
		0o644,
	))

	// Set up crontab mock with entries.
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"crontab -l": {
				Output: "0 2 * * * /usr/local/bin/backup.sh\n*/15 * * * * /usr/local/bin/healthcheck.sh",
			},
		},
	}

	s := NewScheduledScanner(tmpDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)

	// Crontab assertions.
	require.NotNil(t, result.Crontab)
	assert.Len(t, result.Crontab.Entries, 2)

	// LaunchAgents assertions.
	require.NotNil(t, result.LaunchAgents)
	assert.Len(t, result.LaunchAgents.Plists, 1)
	assert.Contains(t, result.LaunchAgents.Plists[0].Source, "com.user.backup.plist")
}
