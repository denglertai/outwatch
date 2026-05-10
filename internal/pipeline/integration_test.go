// Package pipeline tests end-to-end pipeline behavior.
package pipeline

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/denglertai/outwatch/internal/targets"
	"github.com/denglertai/outwatch/internal/targets/logback"
)

// TestProcessor_MergeAndConflictFallback verifies merge and conflict fallback semantics.
func TestProcessor_MergeAndConflictFallback(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("a.yaml", "target: logback\nfilename: app.xml\nlogback:\n  loggers:\n    com.a: INFO\n")
	write("b.yml", "target: logback\nfilename: app.xml\nlogback:\n  loggers:\n    com.b: DEBUG\n")
	write("c.yaml", "target: logback\nfilename: other.xml\nlogback:\n  loggers:\n    com.c: WARN\n")

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	registry := targets.NewRegistry()
	if err := registry.Register(logback.Renderer{}); err != nil {
		t.Fatal(err)
	}
	proc := NewProcessor(inputDir, outputDir, registry, logger, true)
	if err := proc.InitialSync(); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	appXML := mustRead(t, filepath.Join(outputDir, "app.xml"))
	if !strings.Contains(appXML, `name="com.a"`) || !strings.Contains(appXML, `name="com.b"`) {
		t.Fatalf("expected merged loggers in app.xml, got %s", appXML)
	}

	write("b.yml", "target: logback\nfilename: app.xml\nlogback:\n  loggers:\n    com.a: DEBUG\n")
	if err := proc.ApplyChanges([]string{filepath.Join(inputDir, "b.yml")}); err != nil {
		t.Fatalf("apply changes: %v", err)
	}

	appXML = mustRead(t, filepath.Join(outputDir, "app.xml"))
	if !strings.Contains(appXML, `name="com.a" level="INFO"`) {
		t.Fatalf("expected first file to win conflict, got %s", appXML)
	}
}

// TestProcessor_InvalidUpdateKeepsLastValid verifies invalid updates do not clobber generated output.
func TestProcessor_InvalidUpdateKeepsLastValid(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	cfgPath := filepath.Join(inputDir, "a.yaml")
	if err := os.WriteFile(cfgPath, []byte("target: logback\nfilename: app.xml\nlogback:\n  loggers:\n    com.a: INFO\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	registry := targets.NewRegistry()
	if err := registry.Register(logback.Renderer{}); err != nil {
		t.Fatal(err)
	}
	proc := NewProcessor(inputDir, outputDir, registry, slog.New(slog.NewTextHandler(io.Discard, nil)), true)
	if err := proc.InitialSync(); err != nil {
		t.Fatal(err)
	}

	before := mustRead(t, filepath.Join(outputDir, "app.xml"))

	if err := os.WriteFile(cfgPath, []byte("target: logback\nfilename: app.xml\nlogback:\n  loggers:\n    com.a: NOPE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := proc.ApplyChanges([]string{cfgPath}); err != nil {
		t.Fatal(err)
	}

	after := mustRead(t, filepath.Join(outputDir, "app.xml"))
	if before != after {
		t.Fatalf("expected output to remain unchanged after invalid update")
	}
}

// TestProcessor_InputDirectory_NotExist verifies error when input directory doesn't exist.
func TestProcessor_InputDirectory_NotExist(t *testing.T) {
	outputDir := t.TempDir()

	registry := targets.NewRegistry()
	if err := registry.Register(logback.Renderer{}); err != nil {
		t.Fatal(err)
	}
	proc := NewProcessor("/nonexistent/dir", outputDir, registry, slog.New(slog.NewTextHandler(io.Discard, nil)), true)
	if err := proc.InitialSync(); err == nil {
		t.Fatal("expected error for non-existent input directory")
	}
}

// TestProcessor_UnknownTarget verifies invalid target config is skipped.
func TestProcessor_UnknownTarget(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	cfgPath := filepath.Join(inputDir, "a.yaml")
	if err := os.WriteFile(cfgPath, []byte("target: unknowntarget\nfilename: app.xml\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	registry := targets.NewRegistry()
	if err := registry.Register(logback.Renderer{}); err != nil {
		t.Fatal(err)
	}
	proc := NewProcessor(inputDir, outputDir, registry, slog.New(slog.NewTextHandler(io.Discard, nil)), true)
	if err := proc.InitialSync(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "app.xml")); err == nil {
		t.Fatal("expected no output file for invalid target config")
	}
}

// TestProcessor_MultipleValidConfigs verifies multiple configs generating different outputs.
func TestProcessor_MultipleValidConfigs(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("a.yaml", "target: logback\nfilename: a.xml\nlogback:\n  loggers:\n    com.a: DEBUG\n")
	write("b.yaml", "target: logback\nfilename: b.xml\nlogback:\n  loggers:\n    com.b: INFO\n")
	write("c.yaml", "target: logback\nfilename: c.xml\nlogback:\n  loggers:\n    com.c: WARN\n")

	registry := targets.NewRegistry()
	if err := registry.Register(logback.Renderer{}); err != nil {
		t.Fatal(err)
	}
	proc := NewProcessor(inputDir, outputDir, registry, slog.New(slog.NewTextHandler(io.Discard, nil)), true)
	if err := proc.InitialSync(); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	for _, f := range []string{"a.xml", "b.xml", "c.xml"} {
		_, err := os.Stat(filepath.Join(outputDir, f))
		if err != nil {
			t.Fatalf("expected output file %s to exist: %v", f, err)
		}
	}
}

// mustRead reads a file or fails the test.
func mustRead(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
