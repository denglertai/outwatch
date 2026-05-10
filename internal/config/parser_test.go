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

// TestParseFile_File_Not_Found verifies error when file does not exist.
func TestParseFile_File_Not_Found(t *testing.T) {
	path := "/nonexistent/path/config.yaml"

	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "read config") {
		t.Fatalf("expected read error, got %v", err)
	}
}

// TestParseFile_Invalid_YAML verifies error when YAML is malformed.
func TestParseFile_Invalid_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	// Invalid YAML with bad indentation
	content := "target: logback\n  bad indent: value\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parse YAML") {
		t.Fatalf("expected parse error, got %v", err)
	}
}
