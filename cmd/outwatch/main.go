// Package main provides the outwatch CLI entrypoint.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/denglertai/outwatch/internal/pipeline"
	"github.com/denglertai/outwatch/internal/targets"
	"github.com/denglertai/outwatch/internal/targets/logback"
	"github.com/denglertai/outwatch/internal/watch"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// main configures and starts the outwatch process.
func main() {
	var (
		inputDir        string
		outputDir       string
		once            bool
		showVersion     bool
		debounce        time.Duration
		createOutputDir bool
		deleteStale     bool
		logLevel        string
	)

	flag.StringVar(&inputDir, "input-dir", "", "Directory containing YAML config files")
	flag.StringVar(&outputDir, "output-dir", "", "Directory where generated files are written")
	flag.BoolVar(&once, "once", false, "Run one sync and exit")
	flag.BoolVar(&showVersion, "version", false, "Print version information and exit")
	flag.DurationVar(&debounce, "debounce", 300*time.Millisecond, "Debounce duration for batched file events")
	flag.BoolVar(&createOutputDir, "create-output-dir", false, "Create output directory if it does not exist")
	flag.BoolVar(&deleteStale, "delete-stale-outputs", true, "Delete generated outputs that no longer have source configs")
	flag.StringVar(&logLevel, "log-level", "info", "Log level: debug, info, warn, or error")
	flag.Parse()

	if showVersion {
		_, _ = fmt.Printf("outwatch version=%s commit=%s buildDate=%s\n", version, commit, buildDate)
		return
	}

	level, err := parseLogLevel(logLevel)
	if err != nil {
		fatalf("invalid --log-level: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	if inputDir == "" || outputDir == "" {
		fatalf("both --input-dir and --output-dir are required")
	}

	if err := ensureOutputDir(outputDir, createOutputDir); err != nil {
		fatalf("output directory validation failed: %v", err)
	}

	registry := targets.NewRegistry()
	if err := registry.Register(logback.Renderer{}); err != nil {
		fatalf("register targets: %v", err)
	}

	proc := pipeline.NewProcessor(inputDir, outputDir, registry, logger, deleteStale)
	if err := proc.InitialSync(); err != nil {
		fatalf("initial sync failed: %v", err)
	}

	if once {
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("watching directories", "input_dir", inputDir, "output_dir", outputDir)
	if err := watch.Directory(ctx, inputDir, debounce, func(paths []string) {
		if err := proc.ApplyChanges(paths); err != nil {
			logger.Error("apply changes failed", "error", err)
		}
	}); err != nil {
		fatalf("watcher failed: %v", err)
	}
}

// ensureOutputDir validates that the output directory exists and is writable.
func ensureOutputDir(outputDir string, create bool) error {
	info, err := os.Stat(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			if !create {
				return fmt.Errorf("output directory %q does not exist; set --create-output-dir to create it", outputDir)
			}
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", outputDir)
	}

	testFile, err := os.CreateTemp(outputDir, ".outwatch-perm-*.tmp")
	if err != nil {
		return fmt.Errorf("output directory %q is not writable: %w", outputDir, err)
	}
	name := testFile.Name()
	_ = testFile.Close()
	_ = os.Remove(name)
	return nil
}

// fatalf prints an error to stderr and exits with a non-zero status code.
func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// parseLogLevel maps CLI log-level values to slog levels.
func parseLogLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unsupported level %q (supported: debug, info, warn, error)", raw)
	}
}
