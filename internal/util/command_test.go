package util

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockCommandRunner_Run(t *testing.T) {
	tests := []struct {
		name       string
		responses  map[string]MockResponse
		cmd        string
		args       []string
		wantOutput string
		wantErr    bool
	}{
		{
			name: "known command returns predefined output",
			responses: map[string]MockResponse{
				"brew list": {Output: "go\nnode\npython", Err: nil},
			},
			cmd:        "brew",
			args:       []string{"list"},
			wantOutput: "go\nnode\npython",
			wantErr:    false,
		},
		{
			name: "known command with no args",
			responses: map[string]MockResponse{
				"whoami": {Output: "testuser", Err: nil},
			},
			cmd:        "whoami",
			args:       nil,
			wantOutput: "testuser",
			wantErr:    false,
		},
		{
			name: "known command returns error",
			responses: map[string]MockResponse{
				"fail cmd": {Output: "", Err: fmt.Errorf("command failed")},
			},
			cmd:        "fail",
			args:       []string{"cmd"},
			wantOutput: "",
			wantErr:    true,
		},
		{
			name:      "unknown command returns error",
			responses: map[string]MockResponse{},
			cmd:       "unknown",
			args:      []string{"arg"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandRunner{
				Responses: tt.responses,
			}
			ctx := context.Background()
			output, err := mock.Run(ctx, tt.cmd, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantOutput, output)
			}
		})
	}
}

func TestMockCommandRunner_Run_RecordsCalls(t *testing.T) {
	mock := &MockCommandRunner{
		Responses: map[string]MockResponse{
			"git status": {Output: "clean", Err: nil},
			"git diff":   {Output: "", Err: nil},
		},
	}
	ctx := context.Background()

	_, _ = mock.Run(ctx, "git", "status")
	_, _ = mock.Run(ctx, "git", "diff")

	assert.Equal(t, []string{"git status", "git diff"}, mock.Calls)
}

func TestMockCommandRunner_RunLines(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		wantLines []string
		wantErr   bool
	}{
		{
			name:      "splits output by newline and trims empty lines",
			output:    "line1\n\nline2\nline3\n",
			wantLines: []string{"line1", "line2", "line3"},
			wantErr:   false,
		},
		{
			name:      "single line no trailing newline",
			output:    "only",
			wantLines: []string{"only"},
			wantErr:   false,
		},
		{
			name:      "empty output returns empty slice",
			output:    "",
			wantLines: []string{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandRunner{
				Responses: map[string]MockResponse{
					"cmd": {Output: tt.output, Err: nil},
				},
			}
			ctx := context.Background()
			lines, err := mock.RunLines(ctx, "cmd")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantLines, lines)
			}
		})
	}
}

func TestMockCommandRunner_IsInstalled(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]MockResponse
		cmd       string
		want      bool
	}{
		{
			name: "returns true when command exists with no error",
			responses: map[string]MockResponse{
				"brew": {Output: "", Err: nil},
			},
			cmd:  "brew",
			want: true,
		},
		{
			name: "returns false when command has error response",
			responses: map[string]MockResponse{
				"missing": {Output: "", Err: fmt.Errorf("not found")},
			},
			cmd:  "missing",
			want: false,
		},
		{
			name:      "returns false when command not in responses",
			responses: map[string]MockResponse{},
			cmd:       "unknown",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandRunner{
				Responses: tt.responses,
			}
			ctx := context.Background()
			got := mock.IsInstalled(ctx, tt.cmd)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRealCommandRunner_Run(t *testing.T) {
	runner := &RealCommandRunner{}
	ctx := context.Background()

	output, err := runner.Run(ctx, "echo", "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", output)
}

func TestRealCommandRunner_RunLines(t *testing.T) {
	runner := &RealCommandRunner{}
	ctx := context.Background()

	lines, err := runner.RunLines(ctx, "printf", "line1\nline2\nline3")
	require.NoError(t, err)
	assert.Equal(t, []string{"line1", "line2", "line3"}, lines)
}

func TestRealCommandRunner_Run_Error(t *testing.T) {
	runner := &RealCommandRunner{}
	ctx := context.Background()

	_, err := runner.Run(ctx, "ls", "/nonexistent_path_xyz")
	assert.Error(t, err)
}

func TestRealCommandRunner_RunLines_Error(t *testing.T) {
	runner := &RealCommandRunner{}
	ctx := context.Background()

	lines, err := runner.RunLines(ctx, "ls", "/nonexistent_path_xyz")
	assert.Error(t, err)
	assert.Nil(t, lines)
}

func TestMockCommandRunner_RunLines_Error(t *testing.T) {
	mock := &MockCommandRunner{
		Responses: map[string]MockResponse{
			"fail cmd": {Output: "", Err: fmt.Errorf("command failed")},
		},
	}
	ctx := context.Background()

	lines, err := mock.RunLines(ctx, "fail", "cmd")
	assert.Error(t, err)
	assert.Nil(t, lines)
}

func TestSplitLines_EmptyString(t *testing.T) {
	// Test splitLines indirectly via MockCommandRunner.RunLines with empty output
	mock := &MockCommandRunner{
		Responses: map[string]MockResponse{
			"cmd": {Output: "", Err: nil},
		},
	}
	ctx := context.Background()

	lines, err := mock.RunLines(ctx, "cmd")
	require.NoError(t, err)
	assert.Equal(t, []string{}, lines)
	assert.Empty(t, lines)
}

func TestRealCommandRunner_IsInstalled(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{
			name: "echo is always available",
			cmd:  "echo",
			want: true,
		},
		{
			name: "nonexistent tool returns false",
			cmd:  "nonexistent_tool_xyz",
			want: false,
		},
	}

	runner := &RealCommandRunner{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runner.IsInstalled(ctx, tt.cmd)
			assert.Equal(t, tt.want, got)
		})
	}
}
