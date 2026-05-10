package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

// TestParseLogLevel_ValidLevels verifies all valid log levels are parsed correctly.
func TestParseLogLevel_ValidLevels(t *testing.T) {
	cases := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"Debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"", slog.LevelInfo}, // default
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			level, err := parseLogLevel(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if level != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, level)
			}
		})
	}
}

// TestParseLogLevel_InvalidLevel verifies invalid levels are rejected.
func TestParseLogLevel_InvalidLevel(t *testing.T) {
	cases := []string{
		"invalid",
		"trace",
		"fatal",
		"panic",
		"off",
		"unknown",
	}

	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			_, err := parseLogLevel(tc)
			if err == nil {
				t.Fatalf("expected error for %q, got none", tc)
			}
		})
	}
}

// TestParseLogLevel_WithWhitespace verifies leading/trailing whitespace is handled.
func TestParseLogLevel_WithWhitespace(t *testing.T) {
	cases := []string{
		"  debug  ",
		"\tdebug\t",
		"  info  ",
	}

	for _, tc := range cases {
		level, err := parseLogLevel(tc)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc, err)
		}
		if level != slog.LevelDebug && level != slog.LevelInfo {
			t.Errorf("unexpected level for %q: %v", tc, level)
		}
	}
}

// TestEnsureOutputDir_ExistingWritableDir verifies valid existing directory passes.
func TestEnsureOutputDir_ExistingWritableDir(t *testing.T) {
	dir := t.TempDir()

	err := ensureOutputDir(dir, false)
	if err != nil {
		t.Fatalf("unexpected error for existing directory: %v", err)
	}
}

// TestEnsureOutputDir_NonExistentNoCrete verifies error when directory doesn't exist and create=false.
func TestEnsureOutputDir_NonExistentNoCreate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent")

	err := ensureOutputDir(path, false)
	if err == nil {
		t.Fatal("expected error for non-existent directory without create flag")
	}
}

// TestEnsureOutputDir_NonExistentWithCreate verifies directory is created when create=true.
func TestEnsureOutputDir_NonExistentWithCreate(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "new_dir")

	// Verify it doesn't exist
	if _, err := os.Stat(path); err == nil {
		t.Fatal("directory should not exist before test")
	}

	err := ensureOutputDir(path, true)
	if err != nil {
		t.Fatalf("unexpected error creating directory: %v", err)
	}

	// Verify it now exists
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("path exists but is not a directory")
	}
}

// TestEnsureOutputDir_FileNotDirectory verifies error when path is a file, not directory.
func TestEnsureOutputDir_FileNotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := ensureOutputDir(filePath, false)
	if err == nil {
		t.Fatal("expected error when path is a file")
	}
}

// TestEnsureOutputDir_ReadOnlyDirectory verifies error when directory is not writable.
func TestEnsureOutputDir_ReadOnlyDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("skipping permission test as root")
	}

	tmpDir := t.TempDir()

	// Make directory read-only
	if err := os.Chmod(tmpDir, 0o555); err != nil {
		t.Fatalf("failed to chmod directory: %v", err)
	}
	defer os.Chmod(tmpDir, 0o755) // restore for cleanup

	err := ensureOutputDir(tmpDir, false)
	if err == nil {
		t.Fatal("expected error for read-only directory")
	}
}

// TestEnsureOutputDir_NestedPath verifies deeply nested paths are created correctly.
func TestEnsureOutputDir_NestedPath(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "a", "b", "c", "d")

	err := ensureOutputDir(path, true)
	if err != nil {
		t.Fatalf("unexpected error creating nested path: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("nested directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("path exists but is not a directory")
	}
}

// TestEnsureOutputDir_AlreadyExists verifies existing nested paths are handled gracefully.
func TestEnsureOutputDir_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "a", "b", "c")

	// Create it first time
	err := ensureOutputDir(path, true)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	// Create again - should succeed without error
	err = ensureOutputDir(path, true)
	if err != nil {
		t.Fatalf("second create failed: %v", err)
	}
}

// TestEnsureOutputDir_InvalidPath verifies handling of invalid paths.
func TestEnsureOutputDir_InvalidPath(t *testing.T) {
	// On most systems, this path is invalid
	invalidPath := "/dev/null/nonexistent"

	err := ensureOutputDir(invalidPath, true)
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}
