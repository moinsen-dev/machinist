package cloud

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// CloudflareScanner scans Cloudflare Wrangler CLI configuration.
type CloudflareScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewCloudflareScanner creates a new CloudflareScanner with the given homeDir and CommandRunner.
func NewCloudflareScanner(homeDir string, cmd util.CommandRunner) *CloudflareScanner {
	return &CloudflareScanner{
		homeDir: homeDir,
		cmd:     cmd,
	}
}

func (c *CloudflareScanner) Name() string        { return "cloudflare" }
func (c *CloudflareScanner) Description() string { return "Scans Cloudflare Wrangler CLI configuration" }
func (c *CloudflareScanner) Category() string    { return "cloud" }

// Scan checks for wrangler CLI installation, locates config directories, and returns
// a ScanResult with the CloudflareWrangler field populated.
func (c *CloudflareScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: c.Name(),
	}

	if !c.cmd.IsInstalled(ctx, "wrangler") {
		return result, nil
	}

	section := &domain.CloudflareSection{}

	// Primary config location: ~/.config/.wrangler/
	wranglerDir := filepath.Join(c.homeDir, ".config", ".wrangler")
	if util.DirExists(wranglerDir) {
		section.ConfigDir = wranglerDir
	} else {
		// Fallback: ~/.wrangler/config/
		fallbackDir := filepath.Join(c.homeDir, ".wrangler", "config")
		if util.DirExists(fallbackDir) {
			section.ConfigDir = fallbackDir
		}
	}

	result.CloudflareWrangler = section
	return result, nil
}
