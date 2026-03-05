package system

import (
	"context"
	"fmt"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkScanner_Name(t *testing.T) {
	s := NewNetworkScanner(&util.MockCommandRunner{})
	assert.Equal(t, "network", s.Name())
}

func TestNetworkScanner_Description(t *testing.T) {
	s := NewNetworkScanner(&util.MockCommandRunner{})
	assert.Equal(t, "Scans network preferences: Wi-Fi, DNS, and VPN", s.Description())
}

func TestNetworkScanner_Category(t *testing.T) {
	s := NewNetworkScanner(&util.MockCommandRunner{})
	assert.Equal(t, "system", s.Category())
}

func TestNetworkScanner_Scan_HappyPath(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"networksetup -listpreferredwirelessnetworks en0": {
				Output: "Preferred networks on en0:\n\tHomeWifi\n\tOfficeWifi\n\tCoffeeShop",
			},
			"networksetup -getdnsservers Wi-Fi": {
				Output: "8.8.8.8\n8.8.4.4",
			},
			"scutil --nc list": {
				Output: `* (Disabled) XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX IPSec  "Office VPN"
* (Connected) YYYYYYYY-YYYY-YYYY-YYYY-YYYYYYYYYYYY IPSec  "Home VPN"`,
			},
		},
	}

	s := NewNetworkScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Network)

	assert.Equal(t, []string{"HomeWifi", "OfficeWifi", "CoffeeShop"}, result.Network.PreferredWifi)

	require.NotNil(t, result.Network.DNS)
	assert.Equal(t, "Wi-Fi", result.Network.DNS.Interface)
	assert.Equal(t, []string{"8.8.8.8", "8.8.4.4"}, result.Network.DNS.Servers)

	require.Len(t, result.Network.VPNConfigs, 2)
	assert.Equal(t, "Office VPN", result.Network.VPNConfigs[0].Source)
	assert.Equal(t, "Home VPN", result.Network.VPNConfigs[1].Source)
}

func TestNetworkScanner_Scan_NoDNS(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"networksetup -listpreferredwirelessnetworks en0": {
				Output: "Preferred networks on en0:\n\tMyWifi",
			},
			"networksetup -getdnsservers Wi-Fi": {
				Output: "There aren't any DNS Servers set on Wi-Fi.",
			},
			"scutil --nc list": {Err: fmt.Errorf("no VPN")},
		},
	}

	s := NewNetworkScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Network)
	assert.Equal(t, []string{"MyWifi"}, result.Network.PreferredWifi)
	assert.Nil(t, result.Network.DNS)
	assert.Nil(t, result.Network.VPNConfigs)
}

func TestNetworkScanner_Scan_AllErrors(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"networksetup -listpreferredwirelessnetworks en0": {Err: fmt.Errorf("no wifi")},
			"networksetup -getdnsservers Wi-Fi":               {Err: fmt.Errorf("no dns")},
			"scutil --nc list":                                {Err: fmt.Errorf("no vpn")},
		},
	}

	s := NewNetworkScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.Network)
}

func TestNetworkScanner_Scan_OnlyVPN(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"networksetup -listpreferredwirelessnetworks en0": {Err: fmt.Errorf("no wifi")},
			"networksetup -getdnsservers Wi-Fi":               {Err: fmt.Errorf("no dns")},
			"scutil --nc list": {
				Output: `* (Disabled) XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX IPSec  "Corporate VPN"`,
			},
		},
	}

	s := NewNetworkScanner(mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Network)
	assert.Nil(t, result.Network.PreferredWifi)
	assert.Nil(t, result.Network.DNS)
	require.Len(t, result.Network.VPNConfigs, 1)
	assert.Equal(t, "Corporate VPN", result.Network.VPNConfigs[0].Source)
}

func TestParseWifiNetworks(t *testing.T) {
	input := "Preferred networks on en0:\n\tHomeWifi\n\tOfficeWifi"
	expected := []string{"HomeWifi", "OfficeWifi"}
	assert.Equal(t, expected, parseWifiNetworks(input))
}

func TestParseWifiNetworks_Empty(t *testing.T) {
	input := "Preferred networks on en0:"
	assert.Nil(t, parseWifiNetworks(input))
}

func TestParseDNSServers(t *testing.T) {
	input := "8.8.8.8\n8.8.4.4\n1.1.1.1"
	expected := []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"}
	assert.Equal(t, expected, parseDNSServers(input))
}

func TestParseVPNConfigs(t *testing.T) {
	input := `* (Disabled) XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX IPSec  "Office VPN"
* (Connected) YYYYYYYY-YYYY-YYYY-YYYY-YYYYYYYYYYYY IPSec  "Home VPN"`
	configs := parseVPNConfigs(input)
	require.Len(t, configs, 2)
	assert.Equal(t, "Office VPN", configs[0].Source)
	assert.Equal(t, "Home VPN", configs[1].Source)
}

func TestParseVPNConfigs_NoQuotes(t *testing.T) {
	input := "some line without quotes"
	configs := parseVPNConfigs(input)
	assert.Nil(t, configs)
}

func TestExtractQuotedName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`* (Disabled) XXX IPSec  "My VPN"`, "My VPN"},
		{`no quotes here`, ""},
		{`"only start`, ""},
		{`"complete"`, "complete"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, extractQuotedName(tt.input))
	}
}
