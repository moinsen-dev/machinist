package main

import (
	"os"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/scanner/cloud"
	"github.com/moinsen-dev/machinist/internal/scanner/editors"
	gitscanner "github.com/moinsen-dev/machinist/internal/scanner/git"
	"github.com/moinsen-dev/machinist/internal/scanner/packages"
	"github.com/moinsen-dev/machinist/internal/scanner/runtimes"
	"github.com/moinsen-dev/machinist/internal/scanner/security"
	"github.com/moinsen-dev/machinist/internal/scanner/shell"
	"github.com/moinsen-dev/machinist/internal/scanner/system"
	"github.com/moinsen-dev/machinist/internal/scanner/tools"
	"github.com/moinsen-dev/machinist/internal/util"
)

func newRegistry() *scanner.Registry {
	cmd := &util.RealCommandRunner{}
	homeDir, _ := os.UserHomeDir()

	// Default git repo search paths
	searchPaths := []string{
		filepath.Join(homeDir, "Code"),
		filepath.Join(homeDir, "Projects"),
		filepath.Join(homeDir, "Developer"),
		filepath.Join(homeDir, "work"),
	}

	reg := scanner.NewRegistry()
	reg.Register(packages.NewHomebrewScanner(cmd))
	reg.Register(shell.NewShellConfigScanner(homeDir, cmd))
	reg.Register(gitscanner.NewGitReposScanner(searchPaths, cmd))
	reg.Register(runtimes.NewNodeScanner(homeDir, cmd))
	reg.Register(runtimes.NewPythonScanner(cmd))
	reg.Register(runtimes.NewRustScanner(cmd))
	reg.Register(runtimes.NewGoScanner(cmd))
	reg.Register(runtimes.NewJavaScanner(homeDir, cmd))
	reg.Register(runtimes.NewFlutterScanner(cmd))
	reg.Register(runtimes.NewDenoScanner(homeDir, cmd))
	reg.Register(runtimes.NewBunScanner(cmd))
	reg.Register(runtimes.NewRubyScanner(cmd))
	reg.Register(runtimes.NewAsdfScanner(homeDir, cmd))
	reg.Register(editors.NewVSCodeScanner(homeDir, cmd))
	reg.Register(editors.NewCursorScanner(homeDir, cmd))
	reg.Register(editors.NewJetBrainsScanner(homeDir))
	reg.Register(editors.NewNeovimScanner(homeDir, cmd))
	reg.Register(editors.NewXcodeScanner(homeDir, cmd))
	reg.Register(gitscanner.NewGitConfigScanner(homeDir, cmd))
	reg.Register(gitscanner.NewGitHubCLIScanner(homeDir, cmd))
	reg.Register(security.NewSSHScanner(homeDir))
	reg.Register(security.NewGPGScanner(homeDir, cmd))
	reg.Register(shell.NewTerminalScanner(homeDir))
	reg.Register(shell.NewTmuxScanner(homeDir, cmd))
	reg.Register(cloud.NewDockerScanner(homeDir, cmd))
	reg.Register(cloud.NewAWSScanner(homeDir, cmd))
	reg.Register(cloud.NewKubernetesScanner(homeDir, cmd))
	reg.Register(cloud.NewTerraformScanner(homeDir, cmd))
	reg.Register(cloud.NewVercelScanner(homeDir, cmd))
	reg.Register(cloud.NewGCPScanner(homeDir, cmd))
	reg.Register(cloud.NewAzureScanner(homeDir, cmd))
	reg.Register(cloud.NewFlyioScanner(homeDir, cmd))
	reg.Register(cloud.NewFirebaseScanner(homeDir, cmd))
	reg.Register(cloud.NewCloudflareScanner(homeDir, cmd))
	reg.Register(system.NewMacOSDefaultsScanner(cmd))
	reg.Register(system.NewFontsScanner(homeDir, cmd))
	reg.Register(system.NewFoldersScanner(homeDir))
	reg.Register(system.NewScheduledScanner(homeDir, cmd))
	reg.Register(system.NewAppsScanner(cmd))
	reg.Register(system.NewLocaleScanner(cmd))
	reg.Register(system.NewLoginItemsScanner(cmd))
	reg.Register(system.NewHostsFileScanner())
	reg.Register(system.NewNetworkScanner(cmd))
	reg.Register(tools.NewRaycastScanner(homeDir))
	reg.Register(tools.NewAlfredScanner(homeDir))
	reg.Register(tools.NewKarabinerScanner(homeDir))
	reg.Register(tools.NewRectangleScanner(homeDir))
	reg.Register(tools.NewBetterTouchToolScanner(homeDir))
	reg.Register(tools.NewOnePasswordScanner(homeDir, cmd))
	reg.Register(tools.NewDatabasesScanner(homeDir))
	reg.Register(tools.NewRegistriesScanner(homeDir))
	reg.Register(tools.NewBrowserScanner(cmd))
	reg.Register(tools.NewAIToolsScanner(homeDir, cmd))
	reg.Register(tools.NewAPIToolsScanner(homeDir, cmd))
	reg.Register(tools.NewXDGConfigScanner(homeDir))
	reg.Register(tools.NewEnvFilesScanner(homeDir))
	return reg
}
