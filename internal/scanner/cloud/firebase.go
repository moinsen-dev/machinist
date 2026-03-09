package cloud

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// FirebaseScanner scans Firebase CLI configuration.
type FirebaseScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewFirebaseScanner creates a new FirebaseScanner with the given homeDir and CommandRunner.
func NewFirebaseScanner(homeDir string, cmd util.CommandRunner) *FirebaseScanner {
	return &FirebaseScanner{
		homeDir: homeDir,
		cmd:     cmd,
	}
}

func (f *FirebaseScanner) Name() string        { return "firebase" }
func (f *FirebaseScanner) Description() string { return "Scans Firebase CLI configuration" }
func (f *FirebaseScanner) Category() string    { return "cloud" }

// Scan checks for firebase CLI installation, locates config directories, and returns
// a ScanResult with the Firebase field populated.
func (f *FirebaseScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: f.Name(),
	}

	if !f.cmd.IsInstalled(ctx, "firebase") {
		return result, nil
	}

	section := &domain.FirebaseSection{}

	// Primary config location: ~/.config/firebase/
	firebaseDir := filepath.Join(f.homeDir, ".config", "firebase")
	if util.DirExists(firebaseDir) {
		section.ConfigDir = filepath.Join(".config", "firebase")
	} else {
		// Fallback: ~/.config/configstore/firebase-tools.json
		configstoreFile := filepath.Join(f.homeDir, ".config", "configstore", "firebase-tools.json")
		if util.FileExists(configstoreFile) {
			section.ConfigDir = filepath.Join(".config", "configstore")
		}
	}

	if section.ConfigDir != "" {
		result.Firebase = section
	}
	return result, nil
}
