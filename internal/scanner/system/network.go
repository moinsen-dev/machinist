package system

import (
	"context"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// NetworkScanner scans network preferences: Wi-Fi, DNS, and VPN configurations.
type NetworkScanner struct {
	cmd util.CommandRunner
}

// NewNetworkScanner creates a new NetworkScanner with the given CommandRunner.
func NewNetworkScanner(cmd util.CommandRunner) *NetworkScanner {
	return &NetworkScanner{cmd: cmd}
}

func (s *NetworkScanner) Name() string        { return "network" }
func (s *NetworkScanner) Description() string  { return "Scans network preferences: Wi-Fi, DNS, and VPN" }
func (s *NetworkScanner) Category() string     { return "system" }

// Scan reads network settings and returns a ScanResult with the Network field populated.
func (s *NetworkScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	section := &domain.NetworkSection{}
	hasAny := false

	// Preferred Wi-Fi networks
	wifiOutput, err := s.cmd.Run(ctx, "networksetup", "-listpreferredwirelessnetworks", "en0")
	if err == nil {
		networks := parseWifiNetworks(wifiOutput)
		if len(networks) > 0 {
			section.PreferredWifi = networks
			hasAny = true
		}
	}

	// DNS servers
	dnsOutput, err := s.cmd.Run(ctx, "networksetup", "-getdnsservers", "Wi-Fi")
	if err == nil && !strings.Contains(dnsOutput, "There aren't any DNS Servers") {
		servers := parseDNSServers(dnsOutput)
		if len(servers) > 0 {
			section.DNS = &domain.DNSConfig{
				Interface: "Wi-Fi",
				Servers:   servers,
			}
			hasAny = true
		}
	}

	// VPN configurations
	vpnOutput, err := s.cmd.Run(ctx, "scutil", "--nc", "list")
	if err == nil {
		configs := parseVPNConfigs(vpnOutput)
		if len(configs) > 0 {
			section.VPNConfigs = configs
			hasAny = true
		}
	}

	if hasAny {
		result.Network = section
	}

	return result, nil
}

// parseWifiNetworks parses the output of networksetup -listpreferredwirelessnetworks.
// The first line is a header ("Preferred networks on en0:"), subsequent lines are network names.
func parseWifiNetworks(output string) []string {
	lines := strings.Split(output, "\n")
	var networks []string
	for i, line := range lines {
		if i == 0 {
			// Skip header line
			continue
		}
		name := strings.TrimSpace(line)
		if name != "" {
			networks = append(networks, name)
		}
	}
	return networks
}

// parseDNSServers parses DNS server addresses from networksetup output (one per line).
func parseDNSServers(output string) []string {
	var servers []string
	for _, line := range strings.Split(output, "\n") {
		server := strings.TrimSpace(line)
		if server != "" {
			servers = append(servers, server)
		}
	}
	return servers
}

// parseVPNConfigs parses VPN configuration names from scutil --nc list output.
// Each line looks like: * (Disabled) XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX IPSec  "VPN Name"
func parseVPNConfigs(output string) []domain.ConfigFile {
	var configs []domain.ConfigFile
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Extract the quoted name
		name := extractQuotedName(line)
		if name != "" {
			configs = append(configs, domain.ConfigFile{
				Source: name,
			})
		}
	}
	return configs
}

// extractQuotedName extracts a double-quoted string from a line.
func extractQuotedName(line string) string {
	start := strings.Index(line, "\"")
	if start < 0 {
		return ""
	}
	end := strings.Index(line[start+1:], "\"")
	if end < 0 {
		return ""
	}
	return line[start+1 : start+1+end]
}
