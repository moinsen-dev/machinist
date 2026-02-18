package scanner

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/moinsen-dev/machinist/internal/domain"
)

// ScanResult holds the output of a single scanner.
// Each scanner populates exactly one section field.
type ScanResult struct {
	ScannerName string
	Duration    time.Duration
	Err         error

	// Each scanner populates one of these (the rest remain nil):
	Homebrew      *domain.HomebrewSection
	Shell         *domain.ShellSection
	Node          *domain.NodeSection
	Python        *domain.PythonSection
	Rust          *domain.RustSection
	Git           *domain.GitSection
	GitRepos      *domain.GitReposSection
	VSCode        *domain.VSCodeSection
	Cursor        *domain.CursorSection
	Docker        *domain.DockerSection
	MacOSDefaults *domain.MacOSDefaultsSection
	Folders       *domain.FoldersSection
	Fonts         *domain.FontsSection
	Crontab       *domain.CrontabSection
	LaunchAgents  *domain.LaunchAgentsSection
	Apps          *domain.AppsSection
}

// Scanner is the interface that all scanners must implement.
type Scanner interface {
	Name() string
	Description() string
	Category() string
	Scan(ctx context.Context) (*ScanResult, error)
}

// Registry manages all registered scanners.
type Registry struct {
	scanners map[string]Scanner
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		scanners: make(map[string]Scanner),
	}
}

// Register adds a scanner. Returns error if name already registered.
func (r *Registry) Register(s Scanner) error {
	name := s.Name()
	if _, exists := r.scanners[name]; exists {
		return fmt.Errorf("scanner already registered: %s", name)
	}
	r.scanners[name] = s
	return nil
}

// Get returns a scanner by name.
func (r *Registry) Get(name string) (Scanner, error) {
	s, ok := r.scanners[name]
	if !ok {
		return nil, fmt.Errorf("scanner not found: %s", name)
	}
	return s, nil
}

// List returns all registered scanners sorted by name.
func (r *Registry) List() []Scanner {
	scanners := make([]Scanner, 0, len(r.scanners))
	for _, s := range r.scanners {
		scanners = append(scanners, s)
	}
	sort.Slice(scanners, func(i, j int) bool {
		return scanners[i].Name() < scanners[j].Name()
	})
	return scanners
}

// ProgressEvent describes what happened during a scan step.
type ProgressEvent struct {
	Name     string
	Index    int
	Total    int
	Done     bool
	Duration time.Duration
	Err      error
}

// ProgressFunc is called before (Done=false) and after (Done=true) each scanner.
type ProgressFunc func(event ProgressEvent)

// ScanAll runs all registered scanners sequentially and merges results into a Snapshot.
func (r *Registry) ScanAll(ctx context.Context) (*domain.Snapshot, []error) {
	return r.ScanAllWithProgress(ctx, nil)
}

// ScanAllWithProgress runs all scanners with an optional progress callback.
func (r *Registry) ScanAllWithProgress(ctx context.Context, onProgress ProgressFunc) (*domain.Snapshot, []error) {
	start := time.Now()

	hostname, _ := os.Hostname()
	arch := runtime.GOARCH
	osVersion := runtime.GOOS

	snap := domain.NewSnapshot(hostname, osVersion, arch, "")

	scanners := r.List()
	total := len(scanners)
	var errs []error

	for i, s := range scanners {
		if onProgress != nil {
			onProgress(ProgressEvent{Name: s.Name(), Index: i, Total: total})
		}

		scanStart := time.Now()
		result, err := s.Scan(ctx)
		elapsed := time.Since(scanStart)

		if err != nil {
			errs = append(errs, fmt.Errorf("scanner %s: %w", s.Name(), err))
			if onProgress != nil {
				onProgress(ProgressEvent{Name: s.Name(), Index: i, Total: total, Done: true, Duration: elapsed, Err: err})
			}
			continue
		}
		applyResult(snap, result)

		if onProgress != nil {
			onProgress(ProgressEvent{Name: s.Name(), Index: i, Total: total, Done: true, Duration: elapsed})
		}
	}

	snap.Meta.ScanDurationSecs = time.Since(start).Seconds()
	return snap, errs
}

// ScanOne runs a single scanner by name and returns its result.
func (r *Registry) ScanOne(ctx context.Context, name string) (*ScanResult, error) {
	s, err := r.Get(name)
	if err != nil {
		return nil, err
	}
	result, err := s.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("scanner %s: %w", name, err)
	}
	return result, nil
}

// applyResult maps a ScanResult's populated fields onto the Snapshot.
func applyResult(snap *domain.Snapshot, result *ScanResult) {
	if result.Homebrew != nil {
		snap.Homebrew = result.Homebrew
	}
	if result.Shell != nil {
		snap.Shell = result.Shell
	}
	if result.Node != nil {
		snap.Node = result.Node
	}
	if result.Python != nil {
		snap.Python = result.Python
	}
	if result.Rust != nil {
		snap.Rust = result.Rust
	}
	if result.Git != nil {
		snap.Git = result.Git
	}
	if result.GitRepos != nil {
		snap.GitRepos = result.GitRepos
	}
	if result.VSCode != nil {
		snap.VSCode = result.VSCode
	}
	if result.Cursor != nil {
		snap.Cursor = result.Cursor
	}
	if result.Docker != nil {
		snap.Docker = result.Docker
	}
	if result.MacOSDefaults != nil {
		snap.MacOSDefaults = result.MacOSDefaults
	}
	if result.Folders != nil {
		snap.Folders = result.Folders
	}
	if result.Fonts != nil {
		snap.Fonts = result.Fonts
	}
	if result.Crontab != nil {
		snap.Crontab = result.Crontab
	}
	if result.LaunchAgents != nil {
		snap.LaunchAgents = result.LaunchAgents
	}
	if result.Apps != nil {
		snap.Apps = result.Apps
	}
}
