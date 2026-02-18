package bundler

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/util"
)

// BundleOptions holds optional parameters for the Bundle function.
type BundleOptions struct {
	Password        string
	VolumeName      string
	ConfigSourceDir string
}

// PrepareBundleDir creates the bundle directory structure with manifest, install script,
// and config files copied from the source.
func PrepareBundleDir(snapshot *domain.Snapshot, outputDir string, configSourceDir string) error {
	// Create outputDir with configs/ subdirectory
	configsDir := filepath.Join(outputDir, "configs")
	if err := os.MkdirAll(configsDir, 0755); err != nil {
		return fmt.Errorf("create bundle dirs: %w", err)
	}

	// Write manifest.toml
	manifestPath := filepath.Join(outputDir, "manifest.toml")
	if err := domain.WriteManifest(snapshot, manifestPath); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	// Generate and write install.command
	script, err := GenerateRestoreScript(snapshot)
	if err != nil {
		return fmt.Errorf("generate restore script: %w", err)
	}
	installPath := filepath.Join(outputDir, "install.command")
	if err := os.WriteFile(installPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("write install.command: %w", err)
	}

	// Generate and write README.md
	readme, err := GenerateReadme(snapshot)
	if err != nil {
		return fmt.Errorf("generate README: %w", err)
	}
	readmePath := filepath.Join(outputDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return fmt.Errorf("write README.md: %w", err)
	}

	// Generate and write POST_RESTORE_CHECKLIST.md
	checklist, err := GenerateChecklist(snapshot)
	if err != nil {
		return fmt.Errorf("generate checklist: %w", err)
	}
	checklistPath := filepath.Join(outputDir, "POST_RESTORE_CHECKLIST.md")
	if err := os.WriteFile(checklistPath, []byte(checklist), 0644); err != nil {
		return fmt.Errorf("write POST_RESTORE_CHECKLIST.md: %w", err)
	}

	// Copy config files referenced in snapshot sections to configs/
	configFiles := collectConfigFiles(snapshot)
	for _, cf := range configFiles {
		if err := copyConfigFile(cf, outputDir); err != nil {
			return fmt.Errorf("copy config file %s: %w", cf.Source, err)
		}
	}

	return nil
}

// CreateDMG creates a DMG disk image from the given source directory using hdiutil.
// If password is non-empty, the DMG is encrypted with AES-256.
func CreateDMG(ctx context.Context, cmd util.CommandRunner, sourceDir, outputPath, volumeName string, password string) error {
	args := []string{
		"create",
		"-volname", volumeName,
		"-srcfolder", sourceDir,
		"-ov",
		"-format", "UDZO",
	}

	if password != "" {
		args = append(args, "-encryption", "AES-256", "-stdinpass")
	}

	args = append(args, outputPath)

	_, err := cmd.Run(ctx, "hdiutil", args...)
	if err != nil {
		return fmt.Errorf("hdiutil create: %w", err)
	}
	return nil
}

// Bundle orchestrates the full DMG bundling process: prepare bundle dir, create DMG,
// and clean up the temporary directory.
func Bundle(ctx context.Context, cmd util.CommandRunner, snapshot *domain.Snapshot, outputPath string, opts BundleOptions) error {
	// Create a temp dir for the bundle contents
	tmpDir, err := os.MkdirTemp("", "machinist-bundle-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	bundleDir := filepath.Join(tmpDir, "machinist")
	if err := PrepareBundleDir(snapshot, bundleDir, opts.ConfigSourceDir); err != nil {
		return fmt.Errorf("prepare bundle: %w", err)
	}

	volumeName := opts.VolumeName
	if volumeName == "" {
		volumeName = "Machinist Restore"
	}

	if err := CreateDMG(ctx, cmd, bundleDir, outputPath, volumeName, opts.Password); err != nil {
		return err
	}

	return nil
}

// collectConfigFiles gathers all ConfigFile entries from every snapshot section.
func collectConfigFiles(snapshot *domain.Snapshot) []domain.ConfigFile {
	var files []domain.ConfigFile

	if s := snapshot.Shell; s != nil {
		files = append(files, s.ConfigFiles...)
	}
	if s := snapshot.Terminal; s != nil {
		files = append(files, s.ConfigFiles...)
	}
	if s := snapshot.Tmux; s != nil {
		files = append(files, s.ConfigFiles...)
	}
	if s := snapshot.Git; s != nil {
		files = append(files, s.ConfigFiles...)
	}
	if s := snapshot.VSCode; s != nil {
		files = append(files, s.ConfigFiles...)
	}
	if s := snapshot.Cursor; s != nil {
		files = append(files, s.ConfigFiles...)
	}
	if s := snapshot.Xcode; s != nil {
		files = append(files, s.ConfigFiles...)
	}
	if s := snapshot.GPG; s != nil {
		files = append(files, s.ConfigFiles...)
	}
	if s := snapshot.LaunchAgents; s != nil {
		files = append(files, s.Plists...)
	}
	if s := snapshot.Network; s != nil {
		files = append(files, s.VPNConfigs...)
	}
	if s := snapshot.APITools; s != nil {
		files = append(files, s.ConfigFiles...)
	}
	if s := snapshot.Databases; s != nil {
		files = append(files, s.ConfigFiles...)
	}
	if s := snapshot.Registries; s != nil {
		files = append(files, s.ConfigFiles...)
	}

	return files
}

// copyConfigFile copies a single config file into the bundle directory,
// preserving the BundlePath relative structure.
func copyConfigFile(cf domain.ConfigFile, bundleDir string) error {
	if cf.Source == "" || cf.BundlePath == "" {
		return nil
	}

	srcFile, err := os.Open(cf.Source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destPath := filepath.Join(bundleDir, cf.BundlePath)
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	return nil
}
