package bundler

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/security"
	"github.com/moinsen-dev/machinist/internal/util"
)

// BundleOptions holds optional parameters for the Bundle function.
type BundleOptions struct {
	Password        string // DMG-level encryption password (hdiutil AES-256)
	Passphrase      string // age encryption passphrase for sensitive files (SSH keys, GPG, .env)
	VolumeName      string
	ConfigSourceDir string
}

// PrepareBundleDir creates the bundle directory structure with manifest, install script,
// and config files copied from the source. If passphrase is non-empty, sensitive files
// (SSH keys, GPG configs, .env files) are encrypted with age before bundling.
func PrepareBundleDir(snapshot *domain.Snapshot, outputDir string, configSourceDir string, passphrase string) error {
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

	// Generate and write split restore scripts
	scripts, err := GenerateRestoreScripts(snapshot)
	if err != nil {
		return fmt.Errorf("generate restore scripts: %w", err)
	}
	for filename, content := range scripts {
		path := filepath.Join(outputDir, filename)
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			return fmt.Errorf("write %s: %w", filename, err)
		}
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
	if configSourceDir == "" {
		homeDir, _ := os.UserHomeDir()
		configSourceDir = homeDir
	}

	// Emit warnings for sensitive files
	configFiles := collectConfigFiles(snapshot)
	for _, cf := range configFiles {
		if cf.Sensitive {
			fmt.Fprintf(os.Stderr, "Warning: bundling sensitive file: %s\n", cf.Source)
		}
	}

	// Encrypt SSH keys if present and passphrase provided
	if snapshot.SSH != nil && snapshot.SSH.Encrypted && passphrase != "" {
		sshDir := filepath.Join(outputDir, "configs", "ssh")
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			return fmt.Errorf("create ssh bundle dir: %w", err)
		}
		for _, key := range snapshot.SSH.Keys {
			srcPath := filepath.Join(configSourceDir, ".ssh", key)
			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				continue // skip missing key files
			}
			dstPath := filepath.Join(sshDir, key+".age")
			if err := security.EncryptFile(srcPath, dstPath, passphrase); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to encrypt SSH key %s: %v\n", key, err)
				continue
			}
		}
		// Copy non-secret SSH files (config, known_hosts) unencrypted
		if snapshot.SSH.ConfigFile != "" {
			_ = copyConfigFile(domain.ConfigFile{
				Source:     filepath.Join(".ssh", "config"),
				BundlePath: filepath.Join("configs", "ssh", "config"),
			}, outputDir, configSourceDir)
		}
		if snapshot.SSH.KnownHosts != "" {
			_ = copyConfigFile(domain.ConfigFile{
				Source:     filepath.Join(".ssh", "known_hosts"),
				BundlePath: filepath.Join("configs", "ssh", "known_hosts"),
			}, outputDir, configSourceDir)
		}
	}

	// Encrypt GPG config files if present and passphrase provided
	if snapshot.GPG != nil && snapshot.GPG.Encrypted && passphrase != "" {
		gpgDir := filepath.Join(outputDir, "configs", "gpg")
		if err := os.MkdirAll(gpgDir, 0700); err != nil {
			return fmt.Errorf("create gpg bundle dir: %w", err)
		}
		for _, cf := range snapshot.GPG.ConfigFiles {
			srcPath := filepath.Join(configSourceDir, cf.Source)
			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				continue
			}
			baseName := filepath.Base(cf.Source)
			dstPath := filepath.Join(gpgDir, baseName+".age")
			if err := security.EncryptFile(srcPath, dstPath, passphrase); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to encrypt GPG config %s: %v\n", cf.Source, err)
				continue
			}
		}
	}

	// Encrypt .env files if present and passphrase provided
	if snapshot.EnvFiles != nil && snapshot.EnvFiles.Encrypted && passphrase != "" {
		for _, ef := range snapshot.EnvFiles.Files {
			srcPath := filepath.Join(configSourceDir, ef.Source)
			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				continue
			}
			dstPath := filepath.Join(outputDir, ef.BundlePath+".age")
			if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
				return fmt.Errorf("create env file bundle dir: %w", err)
			}
			if err := security.EncryptFile(srcPath, dstPath, passphrase); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to encrypt env file %s: %v\n", ef.Source, err)
				continue
			}
		}
	}

	// Copy remaining (non-encrypted) config files
	for _, cf := range configFiles {
		if err := copyConfigFile(cf, outputDir, configSourceDir); err != nil {
			return fmt.Errorf("copy config file %s: %w", cf.Source, err)
		}
	}

	// Copy config directories (GitHub CLI, Neovim, cloud CLIs, productivity tools, etc.)
	for _, dir := range collectConfigDirs(snapshot) {
		if err := copyConfigDir(dir, outputDir, configSourceDir); err != nil {
			return fmt.Errorf("copy config dir %s: %w", dir.SourceDir, err)
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
	if err := PrepareBundleDir(snapshot, bundleDir, opts.ConfigSourceDir, opts.Passphrase); err != nil {
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
// Sections with []ConfigFile are appended directly. Sections with a single
// ConfigFile string are wrapped into a ConfigFile struct with a derived BundlePath.
func collectConfigFiles(snapshot *domain.Snapshot) []domain.ConfigFile {
	var files []domain.ConfigFile

	// Helper to wrap a single config file path string into a ConfigFile.
	wrap := func(source, bundlePrefix string) domain.ConfigFile {
		return domain.ConfigFile{
			Source:     source,
			BundlePath: filepath.Join("configs", bundlePrefix, filepath.Base(source)),
		}
	}

	// Sections with []ConfigFile
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

	// Sections with a single ConfigFile string
	if s := snapshot.Docker; s != nil && s.ConfigFile != "" {
		files = append(files, wrap(s.ConfigFile, "docker"))
	}
	if s := snapshot.AWS; s != nil && s.ConfigFile != "" {
		files = append(files, wrap(s.ConfigFile, "aws"))
	}
	if s := snapshot.Kubernetes; s != nil && s.ConfigFile != "" {
		files = append(files, wrap(s.ConfigFile, "kubernetes"))
	}
	if s := snapshot.Terraform; s != nil && s.ConfigFile != "" {
		files = append(files, wrap(s.ConfigFile, "terraform"))
	}
	if s := snapshot.Flyio; s != nil && s.ConfigFile != "" {
		files = append(files, wrap(s.ConfigFile, "flyio"))
	}
	if s := snapshot.Rectangle; s != nil && s.ConfigFile != "" {
		files = append(files, wrap(s.ConfigFile, "rectangle"))
	}
	if s := snapshot.BetterTouchTool; s != nil && s.ConfigFile != "" {
		files = append(files, wrap(s.ConfigFile, "bettertouchtool"))
	}
	if s := snapshot.Raycast; s != nil && s.ExportFile != "" {
		files = append(files, wrap(s.ExportFile, "raycast"))
	}
	if s := snapshot.AITools; s != nil && s.ClaudeCodeConfig != "" {
		files = append(files, wrap(s.ClaudeCodeConfig, "ai-tools"))
	}

	// Custom fonts
	if s := snapshot.Fonts; s != nil {
		for _, font := range s.CustomFonts {
			if font.BundlePath != "" {
				files = append(files, wrap(font.BundlePath, "fonts"))
			}
		}
	}

	return files
}

// configDirEntry pairs a source directory path with its bundle destination prefix.
type configDirEntry struct {
	SourceDir   string // relative to home, e.g. ".config/gh"
	BundleDir   string // relative to bundle root, e.g. "configs/github-cli"
}

// collectConfigDirs gathers all ConfigDir entries from sections that store
// their configuration as a directory rather than individual files.
func collectConfigDirs(snapshot *domain.Snapshot) []configDirEntry {
	var dirs []configDirEntry

	add := func(sourceDir, bundlePrefix string) {
		dirs = append(dirs, configDirEntry{
			SourceDir: sourceDir,
			BundleDir: filepath.Join("configs", bundlePrefix),
		})
	}

	if s := snapshot.GitHubCLI; s != nil && s.ConfigDir != "" {
		add(s.ConfigDir, "github-cli")
	}
	if s := snapshot.Neovim; s != nil && s.ConfigDir != "" {
		add(s.ConfigDir, "neovim")
	}
	if s := snapshot.Vercel; s != nil && s.ConfigDir != "" {
		add(s.ConfigDir, "vercel")
	}
	if s := snapshot.GCP; s != nil && s.ConfigDir != "" {
		add(s.ConfigDir, "gcp")
	}
	if s := snapshot.Azure; s != nil && s.ConfigDir != "" {
		add(s.ConfigDir, "azure")
	}
	if s := snapshot.Firebase; s != nil && s.ConfigDir != "" {
		add(s.ConfigDir, "firebase")
	}
	if s := snapshot.CloudflareWrangler; s != nil && s.ConfigDir != "" {
		add(s.ConfigDir, "cloudflare")
	}
	if s := snapshot.Karabiner; s != nil && s.ConfigDir != "" {
		add(s.ConfigDir, "karabiner")
	}
	if s := snapshot.Alfred; s != nil && s.ConfigDir != "" {
		add(s.ConfigDir, "alfred")
	}
	if s := snapshot.OnePassword; s != nil && s.ConfigDir != "" {
		add(s.ConfigDir, "onepassword")
	}
	if s := snapshot.XDGConfig; s != nil {
		if s.ConfigDir != "" {
			add(s.ConfigDir, "xdg-config")
		}
		// Also bundle each auto-detected XDG tool subdirectory individually.
		for _, name := range s.AutoDetected {
			add(filepath.Join(".config", name), filepath.Join("xdg-config", name))
		}
	}
	return dirs
}

// excludedDirNames lists directory names that should never be bundled.
// These are virtual environments, caches, and build artifacts that are
// platform-specific and should be recreated on the target machine.
var excludedDirNames = map[string]bool{
	"virtenv":     true,
	".venv":       true,
	"venv":        true,
	"__pycache__": true,
	"node_modules": true,
	".cache":      true,
}

// excludedFileNames lists files that contain ephemeral credentials or tokens
// that expire and should not be bundled. Users should re-authenticate instead.
var excludedFileNames = map[string]bool{
	"access_tokens.db":  true,
	"credentials.db":    true,
	"cookie_jar":        true,
}

// excludedDirPrefixes lists directory name prefixes to exclude (e.g. legacy_credentials).
var excludedDirPrefixes = []string{
	"legacy_credentials",
}

// shouldExcludeDir returns true if the directory name should be excluded from bundling.
func shouldExcludeDir(name string) bool {
	if excludedDirNames[name] {
		return true
	}
	for _, prefix := range excludedDirPrefixes {
		if name == prefix {
			return true
		}
	}
	return false
}

// copyConfigDir copies an entire directory tree into the bundle directory.
// Source is resolved relative to homeDir. Missing directories are silently skipped.
// Directories matching excludedDirNames and files matching excludedFileNames are skipped.
func copyConfigDir(entry configDirEntry, bundleDir string, homeDir string) error {
	srcDir := filepath.Join(homeDir, entry.SourceDir)
	info, err := os.Stat(srcDir)
	if os.IsNotExist(err) || (err == nil && !info.IsDir()) {
		return nil // skip missing or non-directory
	}
	if err != nil {
		return err
	}

	destDir := filepath.Join(bundleDir, entry.BundleDir)

	return filepath.Walk(srcDir, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil // skip unreadable entries
		}
		// Skip symlinks — they may point outside the tree or be broken.
		if fi.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		name := fi.Name()

		// Skip excluded directories entirely
		if fi.IsDir() && shouldExcludeDir(name) {
			return filepath.SkipDir
		}

		// Skip excluded files (ephemeral credentials/tokens)
		if !fi.IsDir() && excludedFileNames[name] {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(destDir, rel)

		if fi.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		// Skip non-regular files (devices, sockets, etc.)
		if !fi.Mode().IsRegular() {
			return nil
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return nil // skip unreadable files
		}
		defer srcFile.Close()

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		destFile, err := os.Create(dest)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}

// copyConfigFile copies a single config file into the bundle directory,
// preserving the BundlePath relative structure. Source is resolved relative
// to homeDir. Missing files are silently skipped.
func copyConfigFile(cf domain.ConfigFile, bundleDir string, homeDir string) error {
	if cf.Source == "" {
		return nil
	}

	srcPath := filepath.Join(homeDir, cf.Source)
	info, err := os.Stat(srcPath)
	if os.IsNotExist(err) {
		// Gracefully skip missing config files
		return nil
	}
	if err == nil && info.IsDir() {
		// Skip directories — they should be handled by copyConfigDir
		return nil
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Use BundlePath if set, otherwise derive from Source
	bundlePath := cf.BundlePath
	if bundlePath == "" {
		bundlePath = filepath.Join("configs", cf.Source)
	}

	destPath := filepath.Join(bundleDir, bundlePath)
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
