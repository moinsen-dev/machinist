package runtimes

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// JavaScanner scans Java versions via SDKMAN or the system java installation.
type JavaScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewJavaScanner creates a new JavaScanner with the given home directory and CommandRunner.
func NewJavaScanner(homeDir string, cmd util.CommandRunner) *JavaScanner {
	return &JavaScanner{homeDir: homeDir, cmd: cmd}
}

func (j *JavaScanner) Name() string        { return "java" }
func (j *JavaScanner) Description() string { return "Scans Java versions via SDKMAN or system java" }
func (j *JavaScanner) Category() string    { return "runtimes" }

// Scan detects Java installations via SDKMAN (preferred) or falls back to system java.
func (j *JavaScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: j.Name(),
	}

	sdkmanDir := filepath.Join(j.homeDir, ".sdkman")
	if info, err := os.Stat(sdkmanDir); err == nil && info.IsDir() {
		section := j.scanWithSDKMAN(sdkmanDir)
		if section != nil {
			result.Java = section
			return result, nil
		}
	}

	// Fallback: system java.
	section := j.scanSystemJava(ctx)
	if section != nil {
		result.Java = section
	}
	return result, nil
}

// scanWithSDKMAN reads the SDKMAN candidates directory structure for Java versions.
func (j *JavaScanner) scanWithSDKMAN(sdkmanDir string) *domain.JavaSection {
	javaCandidatesDir := filepath.Join(sdkmanDir, "candidates", "java")
	entries, err := os.ReadDir(javaCandidatesDir)
	if err != nil {
		return nil
	}

	var versions []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// "current" is a symlink to the active version — skip it from the list.
		if name == "current" {
			continue
		}
		versions = append(versions, name)
	}

	if len(versions) == 0 {
		return nil
	}

	section := &domain.JavaSection{
		Manager:  "sdkman",
		Versions: versions,
	}

	// Resolve the default version from the "current" symlink.
	currentLink := filepath.Join(javaCandidatesDir, "current")
	if target, err := os.Readlink(currentLink); err == nil {
		section.DefaultVersion = filepath.Base(target)
	}

	// Get JAVA_HOME from the current symlink path.
	if section.DefaultVersion != "" {
		section.JavaHome = filepath.Join(javaCandidatesDir, "current")
	}

	return section
}

// scanSystemJava falls back to `java -version` for machines without SDKMAN.
func (j *JavaScanner) scanSystemJava(ctx context.Context) *domain.JavaSection {
	if !j.cmd.IsInstalled(ctx, "java") {
		return nil
	}

	// `java -version` writes to stderr; our mock returns via stdout.
	output, err := j.cmd.Run(ctx, "java", "-version")
	if err != nil {
		return nil
	}

	version := parseJavaVersion(output)
	if version == "" {
		return nil
	}

	section := &domain.JavaSection{
		Versions:       []string{version},
		DefaultVersion: version,
	}

	// Try to get JAVA_HOME from the environment.
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		section.JavaHome = javaHome
	}

	return section
}

// parseJavaVersion extracts the version from `java -version` output.
// Example: `openjdk version "21.0.1" 2023-10-17` -> "21.0.1"
// Example: `java version "1.8.0_392"` -> "1.8.0_392"
func parseJavaVersion(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		// Look for lines containing a quoted version string.
		start := strings.Index(line, `"`)
		if start == -1 {
			continue
		}
		rest := line[start+1:]
		end := strings.Index(rest, `"`)
		if end == -1 {
			continue
		}
		return rest[:end]
	}
	return ""
}
