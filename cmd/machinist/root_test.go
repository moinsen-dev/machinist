package main

import (
	"bytes"
	"strings"
	"testing"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestVersionCommand(t *testing.T) {
	output, err := executeCommand("version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "machinist") {
		t.Errorf("expected output to contain 'machinist', got: %s", output)
	}
}

func TestListScannersCommand(t *testing.T) {
	output, err := executeCommand("list-scanners")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "homebrew") {
		t.Errorf("expected output to contain 'homebrew', got: %s", output)
	}
	if !strings.Contains(output, "shell") {
		t.Errorf("expected output to contain 'shell', got: %s", output)
	}
}

func TestHelpCommand(t *testing.T) {
	output, err := executeCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "snapshot") {
		t.Errorf("expected output to contain 'snapshot', got: %s", output)
	}
	if !strings.Contains(output, "restore") {
		t.Errorf("expected output to contain 'restore', got: %s", output)
	}
}

func TestScanUnknownScanner(t *testing.T) {
	_, err := executeCommand("scan", "unknown-scanner")
	if err == nil {
		t.Fatal("expected error for unknown scanner, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain 'not found', got: %s", err.Error())
	}
}
