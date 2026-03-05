package tools

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// BrowserScanner scans for the default browser configuration.
type BrowserScanner struct {
	cmd util.CommandRunner
}

// NewBrowserScanner creates a new BrowserScanner.
func NewBrowserScanner(cmd util.CommandRunner) *BrowserScanner {
	return &BrowserScanner{cmd: cmd}
}

func (s *BrowserScanner) Name() string        { return "browser" }
func (s *BrowserScanner) Description() string  { return "Scans default browser configuration" }
func (s *BrowserScanner) Category() string     { return "tools" }

// knownBrowsers maps bundle IDs to human-readable names.
var knownBrowsers = map[string]string{
	"com.apple.safari":            "Safari",
	"com.google.chrome":           "Chrome",
	"org.mozilla.firefox":         "Firefox",
	"com.brave.browser":           "Brave",
	"company.thebrowser.browser":  "Arc",
}

// Scan detects the default browser by parsing the LaunchServices plist.
func (s *BrowserScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	section := &domain.BrowserSection{
		ExtensionsChecklist: "Sign in to browser and sync extensions",
	}

	defaultBrowser := s.detectDefaultBrowser(ctx)
	if defaultBrowser != "" {
		section.Default = defaultBrowser
	}

	result.Browser = section
	return result, nil
}

// detectDefaultBrowser attempts to parse the LaunchServices plist for the HTTPS handler.
func (s *BrowserScanner) detectDefaultBrowser(ctx context.Context) string {
	output, err := s.cmd.Run(ctx, "plutil", "-convert", "json", "-o", "-",
		"$HOME/Library/Preferences/com.apple.LaunchServices/com.apple.launchservices.secure.plist")
	if err != nil {
		return ""
	}

	return parseDefaultBrowser(output)
}

// parseDefaultBrowser extracts the default browser from LaunchServices JSON output.
func parseDefaultBrowser(jsonOutput string) string {
	var data struct {
		LSHandlers []struct {
			LSHandlerURLScheme string `json:"LSHandlerURLScheme"`
			LSHandlerRoleAll   string `json:"LSHandlerRoleAll"`
		} `json:"LSHandlers"`
	}

	if err := json.Unmarshal([]byte(jsonOutput), &data); err != nil {
		return ""
	}

	for _, handler := range data.LSHandlers {
		if strings.EqualFold(handler.LSHandlerURLScheme, "https") {
			bundleID := strings.ToLower(handler.LSHandlerRoleAll)
			if name, ok := knownBrowsers[bundleID]; ok {
				return name
			}
			// Return the bundle ID if we don't have a friendly name
			if bundleID != "" {
				return handler.LSHandlerRoleAll
			}
		}
	}

	return ""
}
