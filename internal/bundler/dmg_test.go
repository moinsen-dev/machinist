package bundler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/security"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareBundleDir(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Taps: []string{"homebrew/core"},
			Formulae: []domain.Package{
				{Name: "git", Version: "2.40"},
			},
			Casks: []domain.Package{
				{Name: "firefox"},
			},
		},
		Shell: &domain.ShellSection{
			DefaultShell: "/bin/zsh",
		},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")

	err := PrepareBundleDir(snap, bundleDir, "", "")
	require.NoError(t, err)

	// Directory created with correct structure
	info, err := os.Stat(bundleDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// manifest.toml exists and contains valid TOML
	manifestPath := filepath.Join(bundleDir, "manifest.toml")
	manifestData, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	assert.True(t, len(manifestData) > 0, "manifest.toml should not be empty")
	// Verify it's valid TOML by unmarshalling
	parsed, err := domain.UnmarshalManifest(manifestData)
	require.NoError(t, err)
	assert.Equal(t, "test-mac", parsed.Meta.SourceHostname)

	// install.command exists and starts with #!/bin/bash
	installPath := filepath.Join(bundleDir, "install.command")
	installData, err := os.ReadFile(installPath)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(installData), "#!/bin/bash"), "install.command should start with #!/bin/bash")
	// Check it is executable
	installInfo, err := os.Stat(installPath)
	require.NoError(t, err)
	assert.True(t, installInfo.Mode().Perm()&0100 != 0, "install.command should be executable")

	// configs/ directory exists
	configsDir := filepath.Join(bundleDir, "configs")
	configsInfo, err := os.Stat(configsDir)
	require.NoError(t, err)
	assert.True(t, configsInfo.IsDir())
}

func TestPrepareBundleDir_WithConfigFiles(t *testing.T) {
	// Create source config files in a temp dir
	configSourceDir := t.TempDir()
	zshrcContent := "# zshrc config\nexport PATH=$HOME/bin:$PATH\n"
	zprofileContent := "# zprofile config\n"
	require.NoError(t, os.WriteFile(filepath.Join(configSourceDir, ".zshrc"), []byte(zshrcContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(configSourceDir, ".zprofile"), []byte(zprofileContent), 0644))

	snap := &domain.Snapshot{
		Meta: newMeta(),
		Shell: &domain.ShellSection{
			DefaultShell: "/bin/zsh",
			ConfigFiles: []domain.ConfigFile{
				{Source: ".zshrc", BundlePath: "configs/.zshrc"},
				{Source: ".zprofile", BundlePath: "configs/.zprofile"},
			},
		},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")

	err := PrepareBundleDir(snap, bundleDir, configSourceDir, "")
	require.NoError(t, err)

	// Config files should be copied to configs/ subdirectory
	copiedZshrc, err := os.ReadFile(filepath.Join(bundleDir, "configs", ".zshrc"))
	require.NoError(t, err)
	assert.Equal(t, zshrcContent, string(copiedZshrc))

	copiedZprofile, err := os.ReadFile(filepath.Join(bundleDir, "configs", ".zprofile"))
	require.NoError(t, err)
	assert.Equal(t, zprofileContent, string(copiedZprofile))
}

func TestPrepareBundleDir_IncludesChecklist(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "git"}},
		},
		SSH: &domain.SSHSection{
			Keys: []string{"id_ed25519"},
		},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")

	err := PrepareBundleDir(snap, bundleDir, "", "")
	require.NoError(t, err)

	// POST_RESTORE_CHECKLIST.md should exist
	checklistPath := filepath.Join(bundleDir, "POST_RESTORE_CHECKLIST.md")
	checklistData, err := os.ReadFile(checklistPath)
	require.NoError(t, err)
	checklistContent := string(checklistData)

	assert.Contains(t, checklistContent, "Post-Restore Checklist")
	assert.Contains(t, checklistContent, "test-mac")
	assert.Contains(t, checklistContent, "macOS Permissions")
	assert.Contains(t, checklistContent, "ssh -T git@github.com")

	// README.md should exist
	readmePath := filepath.Join(bundleDir, "README.md")
	readmeData, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	readmeContent := string(readmeData)

	assert.Contains(t, readmeContent, "Machinist Restore Bundle")
	assert.Contains(t, readmeContent, "test-mac")
	assert.Contains(t, readmeContent, "install.command")
}

func TestCreateDMG(t *testing.T) {
	sourceDir := t.TempDir()
	outputPath := filepath.Join(t.TempDir(), "test.dmg")
	volumeName := "Machinist Restore"

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			fmt.Sprintf("hdiutil create -volname %s -srcfolder %s -ov -format UDZO %s",
				volumeName, sourceDir, outputPath): {Output: "", Err: nil},
		},
	}

	err := CreateDMG(context.Background(), mock, sourceDir, outputPath, volumeName, "")
	require.NoError(t, err)

	// Verify hdiutil was called with correct args
	require.Len(t, mock.Calls, 1)
	call := mock.Calls[0]
	assert.Contains(t, call, "hdiutil create")
	assert.Contains(t, call, "-volname "+volumeName)
	assert.Contains(t, call, "-srcfolder "+sourceDir)
	assert.Contains(t, call, "-ov")
	assert.Contains(t, call, "-format UDZO")
	assert.Contains(t, call, outputPath)
}

func TestCreateDMG_HdiutilError(t *testing.T) {
	sourceDir := t.TempDir()
	outputPath := filepath.Join(t.TempDir(), "test.dmg")
	volumeName := "Machinist Restore"

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			fmt.Sprintf("hdiutil create -volname %s -srcfolder %s -ov -format UDZO %s",
				volumeName, sourceDir, outputPath): {Output: "", Err: fmt.Errorf("hdiutil: resource busy")},
		},
	}

	err := CreateDMG(context.Background(), mock, sourceDir, outputPath, volumeName, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hdiutil: resource busy")
}

func TestCreateDMG_WithPassword(t *testing.T) {
	sourceDir := t.TempDir()
	outputPath := filepath.Join(t.TempDir(), "test.dmg")
	volumeName := "Machinist Restore"
	password := "s3cret"

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			fmt.Sprintf("hdiutil create -volname %s -srcfolder %s -ov -format UDZO -encryption AES-256 -stdinpass %s",
				volumeName, sourceDir, outputPath): {Output: "", Err: nil},
		},
	}

	err := CreateDMG(context.Background(), mock, sourceDir, outputPath, volumeName, password)
	require.NoError(t, err)

	// Verify hdiutil was called with encryption flags
	require.Len(t, mock.Calls, 1)
	call := mock.Calls[0]
	assert.Contains(t, call, "-encryption AES-256")
	assert.Contains(t, call, "-stdinpass")
}

func TestPrepareBundleDir_EncryptsSSHKeys(t *testing.T) {
	configSourceDir := t.TempDir()

	// Create fake SSH key files
	sshDir := filepath.Join(configSourceDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0700))
	keyContent := "-----BEGIN OPENSSH PRIVATE KEY-----\nfake-key-data\n-----END OPENSSH PRIVATE KEY-----\n"
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte(keyContent), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "config"), []byte("Host *\n  AddKeysToAgent yes\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "known_hosts"), []byte("github.com ssh-ed25519 AAAA...\n"), 0644))

	snap := &domain.Snapshot{
		Meta: newMeta(),
		SSH: &domain.SSHSection{
			Encrypted:  true,
			Keys:       []string{"id_ed25519"},
			ConfigFile: "~/.ssh/config",
			KnownHosts: "~/.ssh/known_hosts",
		},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")
	passphrase := "test-passphrase-123"

	err := PrepareBundleDir(snap, bundleDir, configSourceDir, passphrase)
	require.NoError(t, err)

	// SSH key should be encrypted with .age extension
	encryptedKeyPath := filepath.Join(bundleDir, "configs", "ssh", "id_ed25519.age")
	encData, err := os.ReadFile(encryptedKeyPath)
	require.NoError(t, err)
	assert.True(t, security.IsEncrypted(encData), "SSH key should be age-encrypted")

	// Verify decryption roundtrip
	decrypted, err := security.Decrypt(encData, passphrase)
	require.NoError(t, err)
	assert.Equal(t, keyContent, string(decrypted))

	// Non-secret files should be copied unencrypted
	configData, err := os.ReadFile(filepath.Join(bundleDir, "configs", "ssh", "config"))
	require.NoError(t, err)
	assert.Contains(t, string(configData), "Host *")

	knownHostsData, err := os.ReadFile(filepath.Join(bundleDir, "configs", "ssh", "known_hosts"))
	require.NoError(t, err)
	assert.Contains(t, string(knownHostsData), "github.com")
}

func TestPrepareBundleDir_EncryptsEnvFiles(t *testing.T) {
	configSourceDir := t.TempDir()

	// Create fake .env file
	envContent := "DATABASE_URL=postgres://localhost/mydb\nSECRET_KEY=s3cret\n"
	projectDir := filepath.Join(configSourceDir, "projects", "myapp")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, ".env"), []byte(envContent), 0600))

	snap := &domain.Snapshot{
		Meta: newMeta(),
		EnvFiles: &domain.EnvFilesSection{
			Encrypted: true,
			Files: []domain.EnvFile{
				{
					Source:     filepath.Join("projects", "myapp", ".env"),
					BundlePath: filepath.Join("configs", "env", "myapp.env"),
				},
			},
		},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")
	passphrase := "env-pass-456"

	err := PrepareBundleDir(snap, bundleDir, configSourceDir, passphrase)
	require.NoError(t, err)

	// .env file should be encrypted with .age extension appended to BundlePath
	encPath := filepath.Join(bundleDir, "configs", "env", "myapp.env.age")
	encData, err := os.ReadFile(encPath)
	require.NoError(t, err)
	assert.True(t, security.IsEncrypted(encData), ".env file should be age-encrypted")

	// Verify decryption roundtrip
	decrypted, err := security.Decrypt(encData, passphrase)
	require.NoError(t, err)
	assert.Equal(t, envContent, string(decrypted))
}

func TestPrepareBundleDir_EncryptsGPGConfigs(t *testing.T) {
	configSourceDir := t.TempDir()

	// Create fake GPG config files
	gnupgDir := filepath.Join(configSourceDir, ".gnupg")
	require.NoError(t, os.MkdirAll(gnupgDir, 0700))
	gpgConfContent := "keyserver hkps://keys.openpgp.org\n"
	require.NoError(t, os.WriteFile(filepath.Join(gnupgDir, "gpg.conf"), []byte(gpgConfContent), 0600))

	snap := &domain.Snapshot{
		Meta: newMeta(),
		GPG: &domain.GPGSection{
			Encrypted: true,
			Keys:      []string{"ABCD1234"},
			ConfigFiles: []domain.ConfigFile{
				{Source: filepath.Join(".gnupg", "gpg.conf"), BundlePath: filepath.Join("configs", "gpg", "gpg.conf")},
			},
		},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")
	passphrase := "gpg-pass-789"

	err := PrepareBundleDir(snap, bundleDir, configSourceDir, passphrase)
	require.NoError(t, err)

	// GPG config should be encrypted with .age extension
	encPath := filepath.Join(bundleDir, "configs", "gpg", "gpg.conf.age")
	encData, err := os.ReadFile(encPath)
	require.NoError(t, err)
	assert.True(t, security.IsEncrypted(encData), "GPG config should be age-encrypted")

	// Verify decryption roundtrip
	decrypted, err := security.Decrypt(encData, passphrase)
	require.NoError(t, err)
	assert.Equal(t, gpgConfContent, string(decrypted))
}

func TestPrepareBundleDir_NoEncryptionWithoutPassphrase(t *testing.T) {
	configSourceDir := t.TempDir()

	// Create fake SSH key
	sshDir := filepath.Join(configSourceDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("fake-key"), 0600))

	snap := &domain.Snapshot{
		Meta: newMeta(),
		SSH: &domain.SSHSection{
			Encrypted: true,
			Keys:      []string{"id_ed25519"},
		},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")

	// Empty passphrase should skip encryption
	err := PrepareBundleDir(snap, bundleDir, configSourceDir, "")
	require.NoError(t, err)

	// No .age files should exist
	encryptedKeyPath := filepath.Join(bundleDir, "configs", "ssh", "id_ed25519.age")
	_, err = os.Stat(encryptedKeyPath)
	assert.True(t, os.IsNotExist(err), "No .age file should exist when passphrase is empty")
}

func TestPrepareBundleDir_SensitiveFileWarning(t *testing.T) {
	configSourceDir := t.TempDir()

	// Create a sensitive config file
	require.NoError(t, os.WriteFile(filepath.Join(configSourceDir, ".npmrc"), []byte("//registry.npmjs.org/:_authToken=secret\n"), 0600))

	snap := &domain.Snapshot{
		Meta: newMeta(),
		Registries: &domain.RegistriesSection{
			ConfigFiles: []domain.ConfigFile{
				{Source: ".npmrc", BundlePath: "configs/.npmrc", Sensitive: true},
			},
		},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")

	// Capture stderr to verify the warning is emitted
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := PrepareBundleDir(snap, bundleDir, configSourceDir, "")
	require.NoError(t, err)

	w.Close()
	os.Stderr = oldStderr

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	stderrOutput := string(buf[:n])

	assert.Contains(t, stderrOutput, "Warning: bundling sensitive file: .npmrc")
}

func TestCollectConfigFiles_SingleStringFields(t *testing.T) {
	snap := &domain.Snapshot{
		Docker:          &domain.DockerSection{ConfigFile: ".docker/config.json"},
		AWS:             &domain.AWSSection{ConfigFile: ".aws/config"},
		Kubernetes:      &domain.KubernetesSection{ConfigFile: ".kube/config"},
		Terraform:       &domain.TerraformSection{ConfigFile: ".terraformrc"},
		Flyio:           &domain.FlyioSection{ConfigFile: ".fly/config.yml"},
		Rectangle:       &domain.RectangleSection{ConfigFile: "Library/Preferences/com.knollsoft.Rectangle.plist"},
		BetterTouchTool: &domain.BetterTouchToolSection{ConfigFile: "Library/Application Support/BetterTouchTool/btt_data.json"},
		Raycast:         &domain.RaycastSection{ExportFile: "raycast-export.json"},
		// Note: AITools.ClaudeCodeConfig is a directory, handled by collectConfigDirs
	}

	files := collectConfigFiles(snap)

	// Should have exactly 8 files from single-string fields
	assert.Len(t, files, 8)

	// Verify each has a proper BundlePath under configs/
	sources := make(map[string]string)
	for _, f := range files {
		sources[f.Source] = f.BundlePath
		assert.Contains(t, f.BundlePath, "configs/")
	}
	assert.Equal(t, "configs/docker/config.json", sources[".docker/config.json"])
	assert.Equal(t, "configs/aws/config", sources[".aws/config"])
	assert.Equal(t, "configs/kubernetes/config", sources[".kube/config"])
	assert.Equal(t, "configs/raycast/raycast-export.json", sources["raycast-export.json"])
}

func TestCollectConfigFiles_IncludesFonts(t *testing.T) {
	snap := &domain.Snapshot{
		Fonts: &domain.FontsSection{
			CustomFonts: []domain.Font{
				{Name: "FiraCode-Regular", BundlePath: "Library/Fonts/FiraCode-Regular.ttf"},
				{Name: "JetBrainsMono", BundlePath: "Library/Fonts/JetBrainsMono.ttf"},
			},
		},
	}

	files := collectConfigFiles(snap)
	require.Len(t, files, 2)
	assert.Equal(t, "Library/Fonts/FiraCode-Regular.ttf", files[0].Source)
	assert.Equal(t, "configs/fonts/FiraCode-Regular.ttf", files[0].BundlePath)
	assert.Equal(t, "Library/Fonts/JetBrainsMono.ttf", files[1].Source)
	assert.Equal(t, "configs/fonts/JetBrainsMono.ttf", files[1].BundlePath)
}

func TestCollectConfigDirs(t *testing.T) {
	snap := &domain.Snapshot{
		GitHubCLI:          &domain.GitHubCLISection{ConfigDir: ".config/gh"},
		Neovim:             &domain.NeovimSection{ConfigDir: ".config/nvim"},
		Vercel:             &domain.VercelSection{ConfigDir: ".vercel"},
		GCP:                &domain.GCPSection{ConfigDir: ".config/gcloud"},
		Azure:              &domain.AzureSection{ConfigDir: ".azure"},
		Firebase:           &domain.FirebaseSection{ConfigDir: ".config/firebase"},
		CloudflareWrangler: &domain.CloudflareSection{ConfigDir: ".config/.wrangler"},
		Karabiner:          &domain.KarabinerSection{ConfigDir: ".config/karabiner"},
		Alfred:             &domain.AlfredSection{ConfigDir: "Library/Application Support/Alfred"},
		OnePassword:        &domain.OnePasswordSection{ConfigDir: ".config/op"},
		XDGConfig:          &domain.XDGConfigSection{ConfigDir: ".config"},
	}

	dirs := collectConfigDirs(snap)
	assert.Len(t, dirs, 11)

	// Verify bundle dir prefixes
	dirMap := make(map[string]string)
	for _, d := range dirs {
		dirMap[d.SourceDir] = d.BundleDir
	}
	assert.Equal(t, "configs/github-cli", dirMap[".config/gh"])
	assert.Equal(t, "configs/neovim", dirMap[".config/nvim"])
	assert.Equal(t, "configs/gcp", dirMap[".config/gcloud"])
}

func TestCopyConfigDir(t *testing.T) {
	// Create fake source directory with nested files
	homeDir := t.TempDir()
	srcDir := filepath.Join(homeDir, ".config", "gh")
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "hosts"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "config.yml"), []byte("editor: vim\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "hosts", "github.com.yml"), []byte("oauth_token: xxx\n"), 0644))

	bundleDir := t.TempDir()
	entry := configDirEntry{
		SourceDir: ".config/gh",
		BundleDir: "configs/github-cli",
	}

	err := copyConfigDir(entry, bundleDir, homeDir)
	require.NoError(t, err)

	// Verify files were copied
	content, err := os.ReadFile(filepath.Join(bundleDir, "configs", "github-cli", "config.yml"))
	require.NoError(t, err)
	assert.Equal(t, "editor: vim\n", string(content))

	content, err = os.ReadFile(filepath.Join(bundleDir, "configs", "github-cli", "hosts", "github.com.yml"))
	require.NoError(t, err)
	assert.Equal(t, "oauth_token: xxx\n", string(content))
}

func TestCopyConfigDir_MissingDir(t *testing.T) {
	homeDir := t.TempDir()
	bundleDir := t.TempDir()

	entry := configDirEntry{
		SourceDir: ".config/nonexistent",
		BundleDir: "configs/nonexistent",
	}

	// Should silently skip missing directories
	err := copyConfigDir(entry, bundleDir, homeDir)
	assert.NoError(t, err)
}

func TestPrepareBundleDir_CopiesSingleStringConfigFiles(t *testing.T) {
	configSourceDir := t.TempDir()

	// Create fake Docker config
	dockerDir := filepath.Join(configSourceDir, ".docker")
	require.NoError(t, os.MkdirAll(dockerDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dockerDir, "config.json"), []byte(`{"auths":{}}`), 0644))

	snap := &domain.Snapshot{
		Meta:   newMeta(),
		Docker: &domain.DockerSection{ConfigFile: ".docker/config.json"},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")

	err := PrepareBundleDir(snap, bundleDir, configSourceDir, "")
	require.NoError(t, err)

	// Docker config should be in the bundle
	content, err := os.ReadFile(filepath.Join(bundleDir, "configs", "docker", "config.json"))
	require.NoError(t, err)
	assert.Equal(t, `{"auths":{}}`, string(content))
}

func TestPrepareBundleDir_WritesGroupScripts(t *testing.T) {
	snap := &domain.Snapshot{
		Meta:     newMeta(),
		Homebrew: &domain.HomebrewSection{Formulae: []domain.Package{{Name: "git"}}},
		Shell:    &domain.ShellSection{DefaultShell: "/bin/zsh"},
	}

	outputDir := t.TempDir()
	err := PrepareBundleDir(snap, outputDir, t.TempDir(), "")
	require.NoError(t, err)

	// Orchestrator must exist
	_, err = os.Stat(filepath.Join(outputDir, "install.command"))
	assert.NoError(t, err)

	// Group scripts for groups with data
	_, err = os.Stat(filepath.Join(outputDir, "01-homebrew.sh"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(outputDir, "03-configs.sh"))
	assert.NoError(t, err)

	// Groups without data must NOT be written
	_, err = os.Stat(filepath.Join(outputDir, "04-runtimes.sh"))
	assert.True(t, os.IsNotExist(err))

	// Scripts must be executable
	info, _ := os.Stat(filepath.Join(outputDir, "01-homebrew.sh"))
	assert.True(t, info.Mode()&0111 != 0)
}

func TestPrepareBundleDir_CopiesConfigDirs(t *testing.T) {
	configSourceDir := t.TempDir()

	// Create fake GitHub CLI config dir
	ghDir := filepath.Join(configSourceDir, ".config", "gh")
	require.NoError(t, os.MkdirAll(ghDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ghDir, "config.yml"), []byte("editor: vim\n"), 0644))

	snap := &domain.Snapshot{
		Meta:      newMeta(),
		GitHubCLI: &domain.GitHubCLISection{ConfigDir: ".config/gh"},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")

	err := PrepareBundleDir(snap, bundleDir, configSourceDir, "")
	require.NoError(t, err)

	// GH config should be in the bundle
	content, err := os.ReadFile(filepath.Join(bundleDir, "configs", "github-cli", "config.yml"))
	require.NoError(t, err)
	assert.Equal(t, "editor: vim\n", string(content))
}
