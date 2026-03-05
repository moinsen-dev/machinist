package system

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
)

// standardHosts contains the default /etc/hosts entries that should be filtered out.
var standardHosts = map[string]bool{
	"127.0.0.1 localhost":           true,
	"::1 localhost":                 true,
	"255.255.255.255 broadcasthost": true,
}

// HostsFileScanner scans /etc/hosts for custom entries.
type HostsFileScanner struct {
	hostsFilePath string
}

// NewHostsFileScanner creates a new HostsFileScanner that reads from /etc/hosts.
func NewHostsFileScanner() *HostsFileScanner {
	return &HostsFileScanner{hostsFilePath: "/etc/hosts"}
}

func (s *HostsFileScanner) Name() string        { return "hosts-file" }
func (s *HostsFileScanner) Description() string  { return "Scans /etc/hosts for custom entries" }
func (s *HostsFileScanner) Category() string     { return "system" }

// Scan reads the hosts file and returns a ScanResult with the HostsFile field populated.
func (s *HostsFileScanner) Scan(_ context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	entries, err := s.parseHostsFile()
	if err != nil {
		// If the file cannot be read, return empty result (not an error)
		return result, nil
	}

	if len(entries) > 0 {
		result.HostsFile = &domain.HostsFileSection{
			CustomEntries: entries,
		}
	}

	return result, nil
}

// parseHostsFile reads and parses the hosts file, filtering out comments and standard entries.
func (s *HostsFileScanner) parseHostsFile() ([]domain.HostEntry, error) {
	f, err := os.Open(s.hostsFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []domain.HostEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Normalize whitespace for comparison
		normalized := normalizeHostLine(line)
		if standardHosts[normalized] {
			continue
		}

		entry := parseHostLine(line)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// normalizeHostLine normalizes a hosts file line for comparison against standard entries.
func normalizeHostLine(line string) string {
	fields := strings.Fields(line)
	return strings.Join(fields, " ")
}

// parseHostLine parses a single hosts file line into a HostEntry.
func parseHostLine(line string) *domain.HostEntry {
	// Remove inline comments
	if idx := strings.Index(line, "#"); idx >= 0 {
		line = line[:idx]
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}

	return &domain.HostEntry{
		IP:        fields[0],
		Hostnames: fields[1:],
	}
}
