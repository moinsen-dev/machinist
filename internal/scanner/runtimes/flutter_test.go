package runtimes

import (
	"context"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlutterScanner_Name(t *testing.T) {
	s := NewFlutterScanner(&util.MockCommandRunner{})
	assert.Equal(t, "flutter", s.Name())
}

func TestFlutterScanner_Description(t *testing.T) {
	s := NewFlutterScanner(&util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestFlutterScanner_Category(t *testing.T) {
	s := NewFlutterScanner(&util.MockCommandRunner{})
	assert.Equal(t, "runtimes", s.Category())
}

func TestFlutterScanner_Scan_Full(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"flutter": {Output: "", Err: nil}, // IsInstalled check
			"flutter --version": {
				Output: "Flutter 3.19.1 • channel stable • https://github.com/flutter/flutter.git\n" +
					"Framework • revision abc123 (3 weeks ago) • 2024-02-13 10:46:30 -0800\n" +
					"Engine • revision xyz789\n" +
					"Tools • Dart 3.3.0 • DevTools 2.31.1",
			},
			"dart pub global list": {
				Output: "dart_style 2.3.4\nfvm 3.0.1\nmelos 4.0.0\n",
			},
		},
	}

	s := NewFlutterScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Flutter)

	flutter := result.Flutter
	assert.Equal(t, "3.19.1", flutter.Version)
	assert.Equal(t, "stable", flutter.Channel)

	require.Len(t, flutter.DartGlobalPackages, 3)
	assert.Equal(t, "dart_style", flutter.DartGlobalPackages[0])
	assert.Equal(t, "fvm", flutter.DartGlobalPackages[1])
	assert.Equal(t, "melos", flutter.DartGlobalPackages[2])
}

func TestFlutterScanner_Scan_NotInstalled(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "flutter" key absent → IsInstalled returns false
		},
	}

	s := NewFlutterScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Flutter)
}

func TestFlutterScanner_Scan_NoDartGlobalPackages(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"flutter": {Output: "", Err: nil},
			"flutter --version": {
				Output: "Flutter 3.16.5 • channel beta • https://github.com/flutter/flutter.git\n",
			},
			"dart pub global list": {
				Output: "",
			},
		},
	}

	s := NewFlutterScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Flutter)

	assert.Equal(t, "3.16.5", result.Flutter.Version)
	assert.Equal(t, "beta", result.Flutter.Channel)
	assert.Empty(t, result.Flutter.DartGlobalPackages)
}

func TestFlutterScanner_Scan_DartPkgListError(t *testing.T) {
	// dart pub global list failing should not cause scan to fail.
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"flutter": {Output: "", Err: nil},
			"flutter --version": {
				Output: "Flutter 3.19.1 • channel stable • https://github.com/flutter/flutter.git\n",
			},
			// "dart pub global list" absent → RunLines will return error, which is swallowed.
		},
	}

	s := NewFlutterScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Flutter)
	assert.Equal(t, "3.19.1", result.Flutter.Version)
	assert.Equal(t, "stable", result.Flutter.Channel)
	assert.Empty(t, result.Flutter.DartGlobalPackages)
}

func TestParseFlutterVersion(t *testing.T) {
	tests := []struct {
		input           string
		expectedVersion string
		expectedChannel string
	}{
		{
			"Flutter 3.19.1 • channel stable • https://github.com/flutter/flutter.git",
			"3.19.1",
			"stable",
		},
		{
			"Flutter 3.16.5 • channel beta • https://github.com/flutter/flutter.git",
			"3.16.5",
			"beta",
		},
		{
			"Flutter 3.20.0-1.0.pre • channel master • https://github.com/flutter/flutter.git",
			"3.20.0-1.0.pre",
			"master",
		},
		{
			"not flutter output",
			"",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, c := parseFlutterVersion(tt.input)
			assert.Equal(t, tt.expectedVersion, v)
			assert.Equal(t, tt.expectedChannel, c)
		})
	}
}
