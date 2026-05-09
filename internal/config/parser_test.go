// Package config tests configuration parsing and validation behavior.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseFile_Valid verifies parsing and normalization for a valid config.
func TestParseFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")
	content := "target: logback\nfilename: app.xml\nlogback:\n  loggers:\n    com.example: info\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	parsed, err := ParseFile(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if parsed.Output.Config.File != "app.xml" {
		t.Fatalf("unexpected output file: %s", parsed.Output.Config.File)
	}
	if parsed.Output.Config.Target != "logback" {
		t.Fatalf("unexpected target: %s", parsed.Output.Config.Target)
	}
	if parsed.Output.Config.Loggers["com.example"] != "INFO" {
		t.Fatalf("expected normalized INFO level")
	}
}

// TestParseFile_PathOutputRejected verifies filename path segments are rejected.
func TestParseFile_PathOutputRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	content := "target: logback\nfilename: nested/app.xml\nlogback:\n  loggers:\n    com.example: INFO\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ParseFile(path)
	if err == nil || !strings.Contains(err.Error(), "file name only") {
		t.Fatalf("expected basename validation error, got %v", err)
	}
}
