// Package pipeline orchestrates parse, merge, render, and write operations.
package pipeline

import (
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/denglertai/outwatch/internal/config"
	"github.com/denglertai/outwatch/internal/discovery"
	"github.com/denglertai/outwatch/internal/grouping"
	"github.com/denglertai/outwatch/internal/merge"
	"github.com/denglertai/outwatch/internal/output"
	"github.com/denglertai/outwatch/internal/targets"
)

// Processor holds runtime state for reconciling input configs into generated outputs.
type Processor struct {
	inputDir       string
	outputDir      string
	registry       *targets.Registry
	log            *slog.Logger
	state          map[string]config.ParsedFile
	generatedPaths map[string]struct{}
	deleteOutputs  bool
}

// NewProcessor creates a pipeline processor for one input/output directory pair.
func NewProcessor(inputDir, outputDir string, registry *targets.Registry, logger *slog.Logger, deleteOutputs bool) *Processor {
	return &Processor{
		inputDir:       inputDir,
		outputDir:      outputDir,
		registry:       registry,
		log:            logger,
		state:          map[string]config.ParsedFile{},
		generatedPaths: map[string]struct{}{},
		deleteOutputs:  deleteOutputs,
	}
}

// InitialSync performs a full directory scan and reconciles generated outputs.
func (p *Processor) InitialSync() error {
	files, err := discovery.Discover(p.inputDir)
	if err != nil {
		return fmt.Errorf("discover files: %w", err)
	}

	for _, path := range files {
		parsed, err := config.ParseFile(path)
		if err != nil {
			p.log.Warn("invalid config", "path", path, "error", err)
			continue
		}
		p.state[path] = parsed
	}

	return p.reconcile()
}

// ApplyChanges applies a batch of changed paths and reconciles outputs.
func (p *Processor) ApplyChanges(paths []string) error {
	for _, path := range paths {
		if !discovery.IsConfigFile(path) {
			continue
		}

		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				delete(p.state, path)
				continue
			}
			return fmt.Errorf("stat %s: %w", path, err)
		}

		parsed, err := config.ParseFile(path)
		if err != nil {
			// Keep last valid state for this file.
			p.log.Warn("invalid config update", "path", path, "error", err)
			continue
		}
		p.state[path] = parsed
	}

	return p.reconcile()
}

// reconcile computes groups, renders outputs, and handles stale output cleanup.
func (p *Processor) reconcile() error {
	groups := grouping.Build(p.state)
	currentGenerated := map[string]struct{}{}

	for _, group := range groups {
		effective, warnings := merge.Group(group)
		for _, warning := range warnings {
			p.log.Warn("merge warning", "message", warning)
		}

		renderer, err := p.registry.Get(group.Target)
		if err != nil {
			p.log.Error("unsupported target", "target", group.Target, "output", group.File, "error", err)
			continue
		}

		payload, err := renderer.Render(effective)
		if err != nil {
			p.log.Error("render failed", "output", group.File, "target", group.Target, "error", err)
			continue
		}

		outputPath, err := output.ResolvePath(p.outputDir, group.File)
		if err != nil {
			p.log.Error("invalid output filename", "output", group.File, "error", err)
			continue
		}

		if err := output.AtomicWrite(outputPath, payload); err != nil {
			p.log.Error("write failed", "path", outputPath, "error", err)
			continue
		}

		currentGenerated[outputPath] = struct{}{}
	}

	if p.deleteOutputs {
		stale := p.stalePaths(currentGenerated)
		for _, path := range stale {
			if err := output.RemoveIfExists(path); err != nil {
				p.log.Warn("remove stale output failed", "path", path, "error", err)
			}
		}
	}

	p.generatedPaths = currentGenerated
	return nil
}

// stalePaths returns previously generated paths that are no longer produced.
func (p *Processor) stalePaths(current map[string]struct{}) []string {
	paths := make([]string, 0)
	for path := range p.generatedPaths {
		if _, ok := current[path]; !ok {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}
