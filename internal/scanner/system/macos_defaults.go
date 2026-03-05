package system

import (
	"context"
	"strconv"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// customDefault defines a custom macOS default to scan.
type customDefault struct {
	Domain    string
	Key       string
	ValueType string
}

// MacOSDefaultsScanner scans macOS system defaults and preferences.
type MacOSDefaultsScanner struct {
	cmd            util.CommandRunner
	customDefaults []customDefault
}

// NewMacOSDefaultsScanner creates a new MacOSDefaultsScanner with the given CommandRunner.
func NewMacOSDefaultsScanner(cmd util.CommandRunner) *MacOSDefaultsScanner {
	return &MacOSDefaultsScanner{cmd: cmd}
}

func (s *MacOSDefaultsScanner) Name() string        { return "macos-defaults" }
func (s *MacOSDefaultsScanner) Description() string  { return "Scans macOS system defaults and preferences" }
func (s *MacOSDefaultsScanner) Category() string     { return "system" }

// AddCustomDefault registers a custom default to scan.
func (s *MacOSDefaultsScanner) AddCustomDefault(domain, key, valueType string) {
	s.customDefaults = append(s.customDefaults, customDefault{
		Domain:    domain,
		Key:       key,
		ValueType: valueType,
	})
}

// Scan reads macOS defaults and returns a ScanResult with the MacOSDefaults field populated.
func (s *MacOSDefaultsScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	section := &domain.MacOSDefaultsSection{}

	// Read dock settings
	section.Dock = s.scanDock(ctx)

	// Read finder settings
	section.Finder = s.scanFinder(ctx)

	// Read keyboard settings
	section.Keyboard = s.scanKeyboard(ctx)

	// Read screenshots settings
	section.Screenshots = s.scanScreenshots(ctx)

	// Read trackpad settings
	section.Trackpad = s.scanTrackpad(ctx)

	// Read mission control settings
	section.MissionControl = s.scanMissionControl(ctx)

	// Read spotlight settings
	section.Spotlight = s.scanSpotlight(ctx)

	// Read menu bar settings
	section.MenuBar = s.scanMenuBar(ctx)

	// Read custom defaults
	section.Defaults = s.scanCustomDefaults(ctx)

	result.MacOSDefaults = section
	return result, nil
}

// readDefault runs `defaults read <domain> <key>` and returns the output string.
// On error it returns an empty string.
func (s *MacOSDefaultsScanner) readDefault(ctx context.Context, domain, key string) string {
	val, err := s.cmd.Run(ctx, "defaults", "read", domain, key)
	if err != nil {
		return ""
	}
	return val
}

// readBool reads a default and interprets "1" as true, everything else as false.
func (s *MacOSDefaultsScanner) readBool(ctx context.Context, domain, key string) (bool, bool) {
	val := s.readDefault(ctx, domain, key)
	if val == "" {
		return false, false
	}
	return val == "1", true
}

// readInt reads a default and parses it as an integer.
func (s *MacOSDefaultsScanner) readInt(ctx context.Context, domain, key string) (int, bool) {
	val := s.readDefault(ctx, domain, key)
	if val == "" {
		return 0, false
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}
	return n, true
}

// readString reads a default and returns the string value.
func (s *MacOSDefaultsScanner) readString(ctx context.Context, domain, key string) (string, bool) {
	val := s.readDefault(ctx, domain, key)
	if val == "" {
		return "", false
	}
	return val, true
}

func (s *MacOSDefaultsScanner) scanDock(ctx context.Context) *domain.DockConfig {
	dock := &domain.DockConfig{}
	hasAny := false

	if v, ok := s.readBool(ctx, "com.apple.dock", "autohide"); ok {
		dock.AutoHide = v
		hasAny = true
	}
	if v, ok := s.readInt(ctx, "com.apple.dock", "tilesize"); ok {
		dock.TileSize = v
		hasAny = true
	}
	if v, ok := s.readString(ctx, "com.apple.dock", "orientation"); ok {
		dock.Orientation = v
		hasAny = true
	}
	if v, ok := s.readBool(ctx, "com.apple.dock", "magnification"); ok {
		dock.Magnification = v
		hasAny = true
	}
	if v, ok := s.readBool(ctx, "com.apple.dock", "show-recents"); ok {
		dock.ShowRecents = v
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return dock
}

func (s *MacOSDefaultsScanner) scanFinder(ctx context.Context) *domain.FinderConfig {
	finder := &domain.FinderConfig{}
	hasAny := false

	if v, ok := s.readBool(ctx, "com.apple.finder", "ShowPathbar"); ok {
		finder.ShowPathBar = v
		hasAny = true
	}
	if v, ok := s.readBool(ctx, "com.apple.finder", "ShowStatusBar"); ok {
		finder.ShowStatusBar = v
		hasAny = true
	}
	if v, ok := s.readBool(ctx, "com.apple.finder", "AppleShowAllFiles"); ok {
		finder.ShowHidden = v
		hasAny = true
	}
	if v, ok := s.readString(ctx, "com.apple.finder", "FXPreferredViewStyle"); ok {
		finder.DefaultView = v
		hasAny = true
	}
	if v, ok := s.readString(ctx, "com.apple.finder", "FXDefaultSearchScope"); ok {
		finder.DefaultSearchScope = v
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return finder
}

func (s *MacOSDefaultsScanner) scanKeyboard(ctx context.Context) *domain.KeyboardConfig {
	kb := &domain.KeyboardConfig{}
	hasAny := false

	if v, ok := s.readInt(ctx, "NSGlobalDomain", "KeyRepeat"); ok {
		kb.KeyRepeat = v
		hasAny = true
	}
	if v, ok := s.readInt(ctx, "NSGlobalDomain", "InitialKeyRepeat"); ok {
		kb.InitialKeyRepeat = v
		hasAny = true
	}
	if v, ok := s.readBool(ctx, "NSGlobalDomain", "ApplePressAndHoldEnabled"); ok {
		kb.ApplePressAndHoldEnabled = v
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return kb
}

func (s *MacOSDefaultsScanner) scanScreenshots(ctx context.Context) *domain.ScreenshotsConfig {
	ss := &domain.ScreenshotsConfig{}
	hasAny := false

	if v, ok := s.readString(ctx, "com.apple.screencapture", "location"); ok {
		ss.Path = v
		hasAny = true
	}
	if v, ok := s.readString(ctx, "com.apple.screencapture", "type"); ok {
		ss.Format = v
		hasAny = true
	}
	if v, ok := s.readBool(ctx, "com.apple.screencapture", "disable-shadow"); ok {
		ss.DisableShadow = v
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return ss
}

func (s *MacOSDefaultsScanner) scanTrackpad(ctx context.Context) *domain.TrackpadConfig {
	tp := &domain.TrackpadConfig{}
	hasAny := false

	if v, ok := s.readBool(ctx, "com.apple.AppleMultitouchTrackpad", "Clicking"); ok {
		tp.TapToClick = v
		hasAny = true
	}
	if val := s.readDefault(ctx, "NSGlobalDomain", "com.apple.trackpad.scaling"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			tp.TrackingSpeed = f
			hasAny = true
		}
	}

	if !hasAny {
		return nil
	}
	return tp
}

// hotCornerName maps macOS hot corner integer values to human-readable names.
func hotCornerName(val int) string {
	switch val {
	case 2:
		return "Mission Control"
	case 3:
		return "Application Windows"
	case 4:
		return "Desktop"
	case 5:
		return "Start Screen Saver"
	case 6:
		return "Disable Screen Saver"
	case 10:
		return "Put Display to Sleep"
	case 11:
		return "Launchpad"
	case 12:
		return "Notification Center"
	case 13:
		return "Lock Screen"
	default:
		return ""
	}
}

func (s *MacOSDefaultsScanner) scanMissionControl(ctx context.Context) *domain.MissionControlConfig {
	corners := &domain.HotCorners{}
	hasAny := false

	keys := []struct {
		key    string
		setter func(string)
	}{
		{"wvous-tl-corner", func(v string) { corners.TopLeft = v }},
		{"wvous-tr-corner", func(v string) { corners.TopRight = v }},
		{"wvous-bl-corner", func(v string) { corners.BottomLeft = v }},
		{"wvous-br-corner", func(v string) { corners.BottomRight = v }},
	}

	for _, k := range keys {
		if v, ok := s.readInt(ctx, "com.apple.dock", k.key); ok {
			name := hotCornerName(v)
			if name != "" {
				k.setter(name)
				hasAny = true
			}
		}
	}

	if !hasAny {
		return nil
	}
	return &domain.MissionControlConfig{HotCorners: corners}
}

func (s *MacOSDefaultsScanner) scanSpotlight(_ context.Context) *domain.SpotlightConfig {
	return nil // Spotlight exclusions require complex plist parsing
}

func (s *MacOSDefaultsScanner) scanMenuBar(ctx context.Context) *domain.MenuBarConfig {
	mb := &domain.MenuBarConfig{}
	hasAny := false

	if v, ok := s.readString(ctx, "com.apple.menuextra.clock", "DateFormat"); ok {
		mb.ClockFormat = v
		hasAny = true
	}
	if val := s.readDefault(ctx, "com.apple.menuextra.battery", "ShowPercent"); val != "" {
		mb.ShowBatteryPercentage = strings.EqualFold(val, "YES") || val == "1"
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return mb
}

func (s *MacOSDefaultsScanner) scanCustomDefaults(ctx context.Context) []domain.MacDefault {
	if len(s.customDefaults) == 0 {
		return nil
	}

	var defaults []domain.MacDefault
	for _, cd := range s.customDefaults {
		val := s.readDefault(ctx, cd.Domain, cd.Key)
		if val == "" {
			continue
		}
		defaults = append(defaults, domain.MacDefault{
			Domain:    cd.Domain,
			Key:       cd.Key,
			Value:     val,
			ValueType: cd.ValueType,
		})
	}
	return defaults
}
