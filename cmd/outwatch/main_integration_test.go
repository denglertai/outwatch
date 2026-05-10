package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

// Test_ParseLogLevel_AllLevels verifies all log levels parse correctly.
func Test_ParseLogLevel_AllLevels(t *testing.T) {
	cases := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
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

// Test_EnsureOutputDir_AllScenarios tests directory creation and validation.
func Test_EnsureOutputDir_AllScenarios(t *testing.T) {
	t.Run("existing_dir", func(t *testing.T) {
		dir := t.TempDir()
		err := ensureOutputDir(dir, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("nonexistent_no_create", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nonexistent")
		err := ensureOutputDir(path, false)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("nonexistent_with_create", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "new_dir")

		err := ensureOutputDir(path, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify it exists
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("directory not created: %v", err)
		}
	})

	t.Run("nested_path", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "a", "b", "c")

		err := ensureOutputDir(path, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify it exists
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("nested directory not created: %v", err)
		}
	})

	t.Run("file_not_dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file.txt")
		if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		err := ensureOutputDir(filePath, false)
		if err == nil {
			t.Fatal("expected error when path is a file")
		}
	})

	t.Run("permission_denied", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("skipping permission test as root")
		}

		tmpDir := t.TempDir()
		if err := os.Chmod(tmpDir, 0o555); err != nil {
			t.Fatalf("failed to chmod: %v", err)
		}
		defer os.Chmod(tmpDir, 0o755)

		err := ensureOutputDir(tmpDir, false)
		if err == nil {
			t.Fatal("expected error for read-only directory")
		}
	})
}
