package system

import (
	"context"
	"fmt"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMacOSDefaultsScanner_Name(t *testing.T) {
	s := NewMacOSDefaultsScanner(&util.MockCommandRunner{})
	assert.Equal(t, "macos-defaults", s.Name())
}

func TestMacOSDefaultsScanner_Category(t *testing.T) {
	s := NewMacOSDefaultsScanner(&util.MockCommandRunner{})
	assert.Equal(t, "system", s.Category())
}

func TestMacOSDefaultsScanner_Scan_DockSettings(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"defaults read com.apple.dock autohide":     {Output: "1"},
			"defaults read com.apple.dock tilesize":     {Output: "48"},
			"defaults read com.apple.dock orientation":  {Output: "bottom"},
			"defaults read com.apple.dock magnification": {Output: "1"},
			"defaults read com.apple.dock show-recents": {Output: "0"},
			// Finder, Keyboard, Screenshots — return errors so they are skipped
			"defaults read com.apple.finder ShowPathbar":          {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder ShowStatusBar":        {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder AppleShowAllFiles":    {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder FXPreferredViewStyle": {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder FXDefaultSearchScope": {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain KeyRepeat":              {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain InitialKeyRepeat":       {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain ApplePressAndHoldEnabled": {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture location":       {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture type":           {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture disable-shadow": {Err: fmt.Errorf("not set")},
		},
	}

	s := NewMacOSDefaultsScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.MacOSDefaults)
	require.NotNil(t, result.MacOSDefaults.Dock)

	dock := result.MacOSDefaults.Dock
	assert.True(t, dock.AutoHide)
	assert.Equal(t, 48, dock.TileSize)
	assert.Equal(t, "bottom", dock.Orientation)
	assert.True(t, dock.Magnification)
	assert.False(t, dock.ShowRecents)
}

func TestMacOSDefaultsScanner_Scan_FinderSettings(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"defaults read com.apple.finder ShowPathbar":          {Output: "1"},
			"defaults read com.apple.finder ShowStatusBar":        {Output: "1"},
			"defaults read com.apple.finder AppleShowAllFiles":    {Output: "0"},
			"defaults read com.apple.finder FXPreferredViewStyle": {Output: "Nlsv"},
			"defaults read com.apple.finder FXDefaultSearchScope": {Output: "SCcf"},
			// Dock — errors
			"defaults read com.apple.dock autohide":     {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock tilesize":     {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock orientation":  {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock magnification": {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock show-recents": {Err: fmt.Errorf("not set")},
			// Keyboard — errors
			"defaults read NSGlobalDomain KeyRepeat":                {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain InitialKeyRepeat":         {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain ApplePressAndHoldEnabled": {Err: fmt.Errorf("not set")},
			// Screenshots — errors
			"defaults read com.apple.screencapture location":       {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture type":           {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture disable-shadow": {Err: fmt.Errorf("not set")},
		},
	}

	s := NewMacOSDefaultsScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.MacOSDefaults)
	require.NotNil(t, result.MacOSDefaults.Finder)

	finder := result.MacOSDefaults.Finder
	assert.True(t, finder.ShowPathBar)
	assert.True(t, finder.ShowStatusBar)
	assert.False(t, finder.ShowHidden)
	assert.Equal(t, "Nlsv", finder.DefaultView)
	assert.Equal(t, "SCcf", finder.DefaultSearchScope)
}

func TestMacOSDefaultsScanner_Scan_KeyboardSettings(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"defaults read NSGlobalDomain KeyRepeat":                {Output: "2"},
			"defaults read NSGlobalDomain InitialKeyRepeat":         {Output: "15"},
			"defaults read NSGlobalDomain ApplePressAndHoldEnabled": {Output: "0"},
			// Dock — errors
			"defaults read com.apple.dock autohide":     {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock tilesize":     {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock orientation":  {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock magnification": {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock show-recents": {Err: fmt.Errorf("not set")},
			// Finder — errors
			"defaults read com.apple.finder ShowPathbar":          {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder ShowStatusBar":        {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder AppleShowAllFiles":    {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder FXPreferredViewStyle": {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder FXDefaultSearchScope": {Err: fmt.Errorf("not set")},
			// Screenshots — errors
			"defaults read com.apple.screencapture location":       {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture type":           {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture disable-shadow": {Err: fmt.Errorf("not set")},
		},
	}

	s := NewMacOSDefaultsScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.MacOSDefaults)
	require.NotNil(t, result.MacOSDefaults.Keyboard)

	kb := result.MacOSDefaults.Keyboard
	assert.Equal(t, 2, kb.KeyRepeat)
	assert.Equal(t, 15, kb.InitialKeyRepeat)
	assert.False(t, kb.ApplePressAndHoldEnabled)
}

func TestMacOSDefaultsScanner_Scan_ScreenshotsSettings(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"defaults read com.apple.screencapture location":       {Output: "~/Desktop"},
			"defaults read com.apple.screencapture type":           {Output: "png"},
			"defaults read com.apple.screencapture disable-shadow": {Output: "1"},
			// Dock — errors
			"defaults read com.apple.dock autohide":     {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock tilesize":     {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock orientation":  {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock magnification": {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock show-recents": {Err: fmt.Errorf("not set")},
			// Finder — errors
			"defaults read com.apple.finder ShowPathbar":          {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder ShowStatusBar":        {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder AppleShowAllFiles":    {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder FXPreferredViewStyle": {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder FXDefaultSearchScope": {Err: fmt.Errorf("not set")},
			// Keyboard — errors
			"defaults read NSGlobalDomain KeyRepeat":                {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain InitialKeyRepeat":         {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain ApplePressAndHoldEnabled": {Err: fmt.Errorf("not set")},
		},
	}

	s := NewMacOSDefaultsScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.MacOSDefaults)
	require.NotNil(t, result.MacOSDefaults.Screenshots)

	ss := result.MacOSDefaults.Screenshots
	assert.Equal(t, "~/Desktop", ss.Path)
	assert.Equal(t, "png", ss.Format)
	assert.True(t, ss.DisableShadow)
}

func TestMacOSDefaultsScanner_Scan_CustomDefaults(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// All standard reads error out
			"defaults read com.apple.dock autohide":                 {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock tilesize":                 {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock orientation":              {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock magnification":            {Err: fmt.Errorf("not set")},
			"defaults read com.apple.dock show-recents":             {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder ShowPathbar":            {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder ShowStatusBar":          {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder AppleShowAllFiles":      {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder FXPreferredViewStyle":   {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder FXDefaultSearchScope":   {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain KeyRepeat":                {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain InitialKeyRepeat":         {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain ApplePressAndHoldEnabled": {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture location":        {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture type":            {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture disable-shadow":  {Err: fmt.Errorf("not set")},
		},
	}

	s := NewMacOSDefaultsScanner(mock)

	// Add custom defaults to scan
	s.AddCustomDefault("com.example.app", "setting1", "string")
	s.AddCustomDefault("com.example.app", "setting2", "bool")

	// Register responses for custom defaults
	mock.Responses["defaults read com.example.app setting1"] = util.MockResponse{Output: "hello"}
	mock.Responses["defaults read com.example.app setting2"] = util.MockResponse{Output: "1"}

	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.MacOSDefaults)
	require.Len(t, result.MacOSDefaults.Defaults, 2)

	assert.Equal(t, "com.example.app", result.MacOSDefaults.Defaults[0].Domain)
	assert.Equal(t, "setting1", result.MacOSDefaults.Defaults[0].Key)
	assert.Equal(t, "hello", result.MacOSDefaults.Defaults[0].Value)
	assert.Equal(t, "string", result.MacOSDefaults.Defaults[0].ValueType)

	assert.Equal(t, "com.example.app", result.MacOSDefaults.Defaults[1].Domain)
	assert.Equal(t, "setting2", result.MacOSDefaults.Defaults[1].Key)
	assert.Equal(t, "1", result.MacOSDefaults.Defaults[1].Value)
	assert.Equal(t, "bool", result.MacOSDefaults.Defaults[1].ValueType)
}

func TestMacOSDefaultsScanner_Scan_DefaultsReadError(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// Dock: some succeed, some fail
			"defaults read com.apple.dock autohide":     {Output: "1"},
			"defaults read com.apple.dock tilesize":     {Err: fmt.Errorf("domain/key pair not found")},
			"defaults read com.apple.dock orientation":  {Err: fmt.Errorf("domain/key pair not found")},
			"defaults read com.apple.dock magnification": {Err: fmt.Errorf("domain/key pair not found")},
			"defaults read com.apple.dock show-recents": {Err: fmt.Errorf("domain/key pair not found")},
			// Finder: all fail
			"defaults read com.apple.finder ShowPathbar":          {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder ShowStatusBar":        {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder AppleShowAllFiles":    {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder FXPreferredViewStyle": {Err: fmt.Errorf("not set")},
			"defaults read com.apple.finder FXDefaultSearchScope": {Err: fmt.Errorf("not set")},
			// Keyboard: all fail
			"defaults read NSGlobalDomain KeyRepeat":                {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain InitialKeyRepeat":         {Err: fmt.Errorf("not set")},
			"defaults read NSGlobalDomain ApplePressAndHoldEnabled": {Err: fmt.Errorf("not set")},
			// Screenshots: all fail
			"defaults read com.apple.screencapture location":       {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture type":           {Err: fmt.Errorf("not set")},
			"defaults read com.apple.screencapture disable-shadow": {Err: fmt.Errorf("not set")},
		},
	}

	s := NewMacOSDefaultsScanner(mock)
	result, err := s.Scan(context.Background())

	// Scanner should not error even though many defaults read commands failed
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.MacOSDefaults)

	// Dock should still be populated with the values that succeeded
	require.NotNil(t, result.MacOSDefaults.Dock)
	assert.True(t, result.MacOSDefaults.Dock.AutoHide)
	// TileSize should be zero value since read failed
	assert.Equal(t, 0, result.MacOSDefaults.Dock.TileSize)
}
