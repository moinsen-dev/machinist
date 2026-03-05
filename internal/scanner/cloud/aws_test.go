package cloud

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWSScanner_Name(t *testing.T) {
	s := NewAWSScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "aws", s.Name())
}

func TestAWSScanner_Description(t *testing.T) {
	s := NewAWSScanner("/tmp", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestAWSScanner_Category(t *testing.T) {
	s := NewAWSScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "cloud", s.Category())
}

func TestAWSScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "aws" key absent -> IsInstalled returns false
		},
	}

	s := NewAWSScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.AWS)
}

func TestAWSScanner_Scan_Full(t *testing.T) {
	homeDir := t.TempDir()

	// Create ~/.aws/config
	awsDir := filepath.Join(homeDir, ".aws")
	require.NoError(t, os.MkdirAll(awsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(awsDir, "config"), []byte("[default]\nregion = us-east-1"), 0644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"aws":                          {Output: "", Err: nil}, // IsInstalled check
			"aws configure list-profiles": {Output: "default\nproduction\nstaging"},
		},
	}

	s := NewAWSScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.AWS)

	aws := result.AWS
	assert.Equal(t, ".aws/config", aws.ConfigFile)
	assert.Equal(t, []string{"default", "production", "staging"}, aws.Profiles)
}

func TestAWSScanner_Scan_NoConfigFile(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"aws":                          {Output: "", Err: nil},
			"aws configure list-profiles": {Output: "default"},
		},
	}

	s := NewAWSScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.AWS)

	assert.Empty(t, result.AWS.ConfigFile)
	assert.Equal(t, []string{"default"}, result.AWS.Profiles)
}

func TestAWSScanner_Scan_ProfilesError(t *testing.T) {
	homeDir := t.TempDir()

	// Create config file
	awsDir := filepath.Join(homeDir, ".aws")
	require.NoError(t, os.MkdirAll(awsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(awsDir, "config"), []byte("[default]"), 0644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"aws":                          {Output: "", Err: nil},
			"aws configure list-profiles": {Output: "", Err: fmt.Errorf("error listing profiles")},
		},
	}

	s := NewAWSScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.AWS)

	assert.Equal(t, ".aws/config", result.AWS.ConfigFile)
	// Profiles error handled gracefully
	assert.Nil(t, result.AWS.Profiles)
}
