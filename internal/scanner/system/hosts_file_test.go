package system

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostsFileScanner_Name(t *testing.T) {
	s := NewHostsFileScanner()
	assert.Equal(t, "hosts-file", s.Name())
}

func TestHostsFileScanner_Description(t *testing.T) {
	s := NewHostsFileScanner()
	assert.Equal(t, "Scans /etc/hosts for custom entries", s.Description())
}

func TestHostsFileScanner_Category(t *testing.T) {
	s := NewHostsFileScanner()
	assert.Equal(t, "system", s.Category())
}

func TestHostsFileScanner_Scan_CustomEntries(t *testing.T) {
	content := `##
# Host Database
#
# localhost is used to configure the loopback interface
##
127.0.0.1	localhost
255.255.255.255	broadcasthost
::1             localhost
# Custom entries
192.168.1.100	myserver.local myserver
10.0.0.1	devbox.internal
`

	tmpFile := writeTempHostsFile(t, content)
	s := &HostsFileScanner{hostsFilePath: tmpFile}
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.HostsFile)
	require.Len(t, result.HostsFile.CustomEntries, 2)

	assert.Equal(t, "192.168.1.100", result.HostsFile.CustomEntries[0].IP)
	assert.Equal(t, []string{"myserver.local", "myserver"}, result.HostsFile.CustomEntries[0].Hostnames)

	assert.Equal(t, "10.0.0.1", result.HostsFile.CustomEntries[1].IP)
	assert.Equal(t, []string{"devbox.internal"}, result.HostsFile.CustomEntries[1].Hostnames)
}

func TestHostsFileScanner_Scan_OnlyStandardEntries(t *testing.T) {
	content := `127.0.0.1	localhost
255.255.255.255	broadcasthost
::1             localhost
`

	tmpFile := writeTempHostsFile(t, content)
	s := &HostsFileScanner{hostsFilePath: tmpFile}
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	// No custom entries, so HostsFile should be nil
	assert.Nil(t, result.HostsFile)
}

func TestHostsFileScanner_Scan_EmptyFile(t *testing.T) {
	tmpFile := writeTempHostsFile(t, "")
	s := &HostsFileScanner{hostsFilePath: tmpFile}
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.HostsFile)
}

func TestHostsFileScanner_Scan_FileNotFound(t *testing.T) {
	s := &HostsFileScanner{hostsFilePath: "/nonexistent/hosts"}
	result, err := s.Scan(context.Background())

	require.NoError(t, err) // Should not error, just return empty
	assert.Nil(t, result.HostsFile)
}

func TestHostsFileScanner_Scan_InlineComments(t *testing.T) {
	content := `127.0.0.1	localhost
192.168.1.50	myapp.local # development server
`

	tmpFile := writeTempHostsFile(t, content)
	s := &HostsFileScanner{hostsFilePath: tmpFile}
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.HostsFile)
	require.Len(t, result.HostsFile.CustomEntries, 1)
	assert.Equal(t, "192.168.1.50", result.HostsFile.CustomEntries[0].IP)
	assert.Equal(t, []string{"myapp.local"}, result.HostsFile.CustomEntries[0].Hostnames)
}

func TestHostsFileScanner_Scan_MultipleHostnames(t *testing.T) {
	content := `127.0.0.1	localhost
10.0.0.5	api.local web.local admin.local
`

	tmpFile := writeTempHostsFile(t, content)
	s := &HostsFileScanner{hostsFilePath: tmpFile}
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.HostsFile)
	require.Len(t, result.HostsFile.CustomEntries, 1)
	assert.Equal(t, "10.0.0.5", result.HostsFile.CustomEntries[0].IP)
	assert.Equal(t, []string{"api.local", "web.local", "admin.local"}, result.HostsFile.CustomEntries[0].Hostnames)
}

func writeTempHostsFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}
