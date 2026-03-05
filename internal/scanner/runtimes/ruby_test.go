package runtimes

import (
	"context"
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRubyScanner_Name(t *testing.T) {
	s := NewRubyScanner(&util.MockCommandRunner{})
	assert.Equal(t, "ruby", s.Name())
}

func TestRubyScanner_Description(t *testing.T) {
	s := NewRubyScanner(&util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestRubyScanner_Category(t *testing.T) {
	s := NewRubyScanner(&util.MockCommandRunner{})
	assert.Equal(t, "runtimes", s.Category())
}

func TestRubyScanner_Scan_WithRbenv(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"rbenv": {Output: "", Err: nil}, // IsInstalled check
			"rbenv versions --bare": {
				Output: "2.7.8\n3.1.4\n3.2.2",
			},
			"rbenv global": {Output: "3.2.2"},
			"gem list --no-versions": {
				Output: "*** LOCAL GEMS ***\nbundler\nrails\nrspec\nrubocop",
			},
		},
	}

	s := NewRubyScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Ruby)

	ruby := result.Ruby
	assert.Equal(t, "rbenv", ruby.Manager)
	assert.Equal(t, "3.2.2", ruby.DefaultVersion)
	assert.Equal(t, []string{"2.7.8", "3.1.4", "3.2.2"}, ruby.Versions)

	require.Len(t, ruby.GlobalGems, 4)
	assert.Equal(t, domain.Package{Name: "bundler"}, ruby.GlobalGems[0])
	assert.Equal(t, domain.Package{Name: "rails"}, ruby.GlobalGems[1])
	assert.Equal(t, domain.Package{Name: "rspec"}, ruby.GlobalGems[2])
	assert.Equal(t, domain.Package{Name: "rubocop"}, ruby.GlobalGems[3])
}

func TestRubyScanner_Scan_WithRvm(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// rbenv absent (no key) → rvm used instead
			"rvm": {Output: "", Err: nil},
			"rvm list strings": {
				Output: "ruby-3.0.6\nruby-3.2.2",
			},
			"rvm current": {Output: "ruby-3.2.2"},
			"gem list --no-versions": {
				Output: "bundler\nrake",
			},
		},
	}

	s := NewRubyScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Ruby)

	ruby := result.Ruby
	assert.Equal(t, "rvm", ruby.Manager)
	assert.Equal(t, "ruby-3.2.2", ruby.DefaultVersion)
	assert.Equal(t, []string{"ruby-3.0.6", "ruby-3.2.2"}, ruby.Versions)

	require.Len(t, ruby.GlobalGems, 2)
	assert.Equal(t, domain.Package{Name: "bundler"}, ruby.GlobalGems[0])
	assert.Equal(t, domain.Package{Name: "rake"}, ruby.GlobalGems[1])
}

func TestRubyScanner_Scan_SystemRuby(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// Neither rbenv nor rvm present.
			"ruby --version": {
				Output: "ruby 3.2.2 (2023-03-30 revision e51014f9c0) [arm64-darwin22]",
			},
			"gem list --no-versions": {
				Output: "bigdecimal\nbundler\ncgi\ncsv",
			},
		},
	}

	s := NewRubyScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Ruby)

	ruby := result.Ruby
	assert.Equal(t, "", ruby.Manager)
	assert.Equal(t, "3.2.2", ruby.DefaultVersion)
	assert.Equal(t, []string{"3.2.2"}, ruby.Versions)
	assert.Len(t, ruby.GlobalGems, 4)
}

func TestRubyScanner_Scan_NoRuby(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// No rbenv, rvm, or ruby keys present.
		},
	}

	s := NewRubyScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Ruby)
}

func TestRubyScanner_Scan_RbenvNoGems(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"rbenv": {Output: "", Err: nil},
			"rbenv versions --bare": {Output: "3.2.2"},
			"rbenv global":          {Output: "3.2.2"},
			// gem list returns empty
			"gem list --no-versions": {Output: ""},
		},
	}

	s := NewRubyScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Ruby)

	assert.Equal(t, "rbenv", result.Ruby.Manager)
	assert.Equal(t, "3.2.2", result.Ruby.DefaultVersion)
	assert.Equal(t, []string{"3.2.2"}, result.Ruby.Versions)
	assert.Empty(t, result.Ruby.GlobalGems)
}

func TestRubyScanner_Scan_GemListHeaderSkipped(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"rbenv": {Output: "", Err: nil},
			"rbenv versions --bare": {Output: "3.2.2"},
			"rbenv global":          {Output: "3.2.2"},
			"gem list --no-versions": {
				Output: "*** LOCAL GEMS ***\nbundler\nrake",
			},
		},
	}

	s := NewRubyScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Ruby)

	gems := result.Ruby.GlobalGems
	require.Len(t, gems, 2)
	assert.Equal(t, "bundler", gems[0].Name)
	assert.Equal(t, "rake", gems[1].Name)
}
