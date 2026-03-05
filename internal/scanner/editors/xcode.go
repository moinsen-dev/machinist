package editors

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// simctlOutput is the top-level structure returned by `xcrun simctl list devices available -j`.
type simctlOutput struct {
	Devices map[string][]struct {
		Name  string `json:"name"`
		State string `json:"state"`
	} `json:"devices"`
}

// XcodeScanner scans Xcode installations, available simulators, and config files.
type XcodeScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewXcodeScanner creates a new XcodeScanner for the given homeDir.
func NewXcodeScanner(homeDir string, cmd util.CommandRunner) *XcodeScanner {
	return &XcodeScanner{homeDir: homeDir, cmd: cmd}
}

func (s *XcodeScanner) Name() string        { return "xcode" }
func (s *XcodeScanner) Description() string { return "Scans Xcode and simulators" }
func (s *XcodeScanner) Category() string    { return "editors" }

// Scan checks for xcode-select, lists available simulators via xcrun simctl, and
// records the Xcode preferences plist if present.
func (s *XcodeScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	if !s.cmd.IsInstalled(ctx, "xcode-select") {
		return result, nil
	}

	section := &domain.XcodeSection{}

	// List available simulators.
	simJSON, err := s.cmd.Run(ctx, "xcrun", "simctl", "list", "devices", "available", "-j")
	if err == nil {
		section.Simulators = parseSimulators(simJSON)
	}

	// Check for Xcode preferences plist.
	plistPath := filepath.Join(s.homeDir, "Library", "Preferences", "com.apple.dt.Xcode.plist")
	if util.FileExists(plistPath) {
		section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
			Source:     "Library/Preferences/com.apple.dt.Xcode.plist",
			BundlePath: "configs/xcode/com.apple.dt.Xcode.plist",
		})
	}

	result.Xcode = section
	return result, nil
}

// parseSimulators extracts unique device names from `xcrun simctl list devices available -j` output.
func parseSimulators(jsonData string) []string {
	var out simctlOutput
	if err := json.Unmarshal([]byte(jsonData), &out); err != nil {
		return nil
	}

	seen := make(map[string]bool)
	var names []string
	for _, devices := range out.Devices {
		for _, device := range devices {
			if device.Name != "" && !seen[device.Name] {
				seen[device.Name] = true
				names = append(names, device.Name)
			}
		}
	}
	return names
}
