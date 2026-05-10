# outwatch Development Instructions

## Project Overview

**outwatch** is a production-ready Go utility that watches a directory of YAML configuration files and generates structured outputs via an extensible target renderer architecture.

**Primary Use Case:** Transform mounted ConfigMaps or filesystem config files into application-specific formats (starting with Logback XML for dynamic log level management).

## What Has Been Built (v0.2.0)

### Core Runtime
- **Config Parser** (`internal/config/`): YAML parsing with strict validation, schema enforcement, source path tracking
- **File Discovery** (`internal/discovery/`): Non-recursive directory scan for .yaml/.yml files
- **Grouping** (`internal/grouping/`): Cluster configs by (target, output-filename) for collision detection
- **Merge** (`internal/merge/`): Resolve multi-file collisions with first-file-wins strategy + warnings
- **Output Writing** (`internal/output/`): Atomic write pattern (temp + rename) with basename-only validation
- **Pipeline** (`internal/pipeline/`): Central orchestrator coordinating parse→group→merge→render→write flow
- **Watcher** (`internal/watch/`): Directory monitoring with fsnotify, debounce timer, sorted batch emission
- **Target System** (`internal/targets/`): Plugin registry + Renderer interface for extensibility

### Logback Target
- Generates valid Logback XML from config
- Deterministic sorted logger output
- Supports per-logger and root logger level configuration
- Fully tested with edge cases (duplicates, missing sections, invalid levels)

### Observability & Operations
- Structured logging via `slog` with configurable level (debug/info/warn/error)
- `--version` flag with build metadata (version, commit, buildDate injected by release workflow)
- `--log-level` CLI flag for runtime control
- `--once` mode for one-shot processing
- `--create-output-dir` auto-creation flag
- `--delete-stale-outputs` cleanup flag

### Deployment Infrastructure
- **Docker:** Multi-stage Dockerfile (golang:1.26-alpine builder → gcr.io/distroless/static-debian12:nonroot)
- **.dockerignore:** Build hygiene (ignores test files, examples, git artifacts)
- **Kubernetes:** Sidecar deployment pattern with ConfigMap inputs and shared volume for outputs
- **Docker Compose:** Local development example with mounted config and volume
- **Makefile:** Targets for test, build, clean, image-build

### Automation
- **CI Workflow** (.github/workflows/ci.yml): Tests on every push/PR
- **Release Workflow** (.github/workflows/release.yml): 
  - Tag-triggered (v*.*.* pattern)
  - Multi-platform binaries: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
  - Windows assets as .zip (not tar.gz)
  - Docker image published to GHCR with version tag + `latest` tag
  - Release notes auto-generated
  - Build metadata injected via ldflags
- **Renovate:** Weekly dependency updates, non-major grouped, no automerge

### Testing & Quality
- Unit tests on all substantive packages (config, grouping, merge, output, targets, registry)
- Integration test (pipeline with multi-file merge + conflict handling)
- End-to-end watcher test (file creation/deletion event batching)
- All tests passing with 0 failures
- GoDoc comments on all public declarations (100% coverage)
- Code formatted with gofmt

## Architecture Patterns

### Config Schema
Single-output-per-file model with target-specific sections:
```yaml
target: logback                    # Required: selects renderer from registry
filename: dynamic-loggers.xml      # Required: basename only (no path separators)
logback:                           # Optional: target-specific config block
  loggers:
    com.example: INFO
    org.hibernate.SQL: DEBUG
```

### Collision Resolution
- Configs with same (target, filename) pair are merged if compatible
- Compatible = same logger definitions or no conflict on logger levels
- Incompatible = conflicting logger levels → fallback to first file, emit warning
- Merge result is written once per group to output file

### Extensibility via Renderer Interface
```go
type Renderer interface {
    Name() string                                    // Unique identifier
    Render(cfg OutputConfig) ([]byte, error)       // Generate output
}
```
New targets added by:
1. Create `internal/targets/mytarget/render.go`
2. Implement `Renderer` interface
3. Register in `main.go`: `registry.Register(myrenderer)`
4. Update config schema in `internal/config/schema.go` with new section type

### Watch Loop Behavior
- Debounce window: coalescesmultiple filesystem events into single batch
- Batch processing: runs full pipeline for all changed files in single batch
- Single-file resilience: invalid/unparseable file doesn't block others in batch
- Delete-stale mode: removes output files whose inputs were all deleted

## Requirements for Working on This Project

### Prerequisites
- Go 1.26+
- Docker + Docker Compose (for image builds and local testing)
- kubectl (for Kubernetes testing, optional)
- Make (for build commands)

### Before Starting Any Work
1. **Run tests:** `make test` or `go test ./...` — all must pass
2. **Format code:** `gofmt -w ./...` — maintain consistency
3. **Understand collision merge:** Read `internal/merge/merge.go` comments; this is non-obvious
4. **Check registry extensibility:** Study `internal/targets/target.go` + `registry.go` patterns

### Adding New Targets
1. Create package `internal/targets/mytarget/`
2. Implement `Renderer` interface in `render.go`
3. Add `mytarget:` section to `FileConfig` struct in `schema.go`
4. Add validation rules for new section in `FileConfig.Validate()`
5. Register in `cmd/outwatch/main.go` main() function
6. Add unit tests in `render_test.go` covering happy path + edge cases
7. Update README with config example + output sample

### Modifying Config Schema
⚠️ Schema changes ripple through: parser → grouping → merge → pipeline → CLI → tests

When updating `internal/config/schema.go`:
1. Update struct tags and validation
2. Update `parser.go` ParseFile() to handle new fields
3. Update `merge.go` Group() merge logic if collision behavior changed
4. Update pipeline integration test (`internal/pipeline/integration_test.go`)
5. Update all affected target tests
6. Update README with new config example

### Testing Requirements
- **Unit tests:** Must exist for all public functions in substantial packages
- **Integration test:** Must verify full pipeline (parse→group→merge→render→write)
- **Edge cases:** Test invalid config, missing fields, collision scenarios, empty input
- **Run before commit:** `go test ./... && gofmt -w ./...` must complete cleanly

### Release Checklist
1. Ensure `go test ./...` passes locally
2. Commit changes with conventional commit (feat:, fix:, docs:, etc.)
3. Create annotated tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z: ..."`
4. Push: `git push origin main vX.Y.Z`
5. GitHub Actions automatically:
   - Runs tests
   - Builds multi-platform binaries
   - Publishes image to GHCR
   - Creates GitHub Release with assets
6. Verify on GitHub: Releases tab shows assets, ghcr.io shows image tags

### Local Development Workflow
```bash
# Run tests
make test

# Build binary
make build
./dist/outwatch --input-dir ./test-configs --output-dir ./outputs --log-level debug

# Build Docker image
make image-build

# Run Docker Compose example
cd examples/docker
docker-compose up
```

### Code Style & Conventions
- Package comments explain purpose (e.g., "// Package watch implements directory monitoring with debounce.")
- Function/struct comments use verb-noun pattern: "Render generates..." not "This generates..."
- Error handling: explicit error checks, meaningful context in errors
- Logging: use slog for structured output, include relevant context (file path, target, etc.)
- Test naming: `TestFunctionName_Scenario` (e.g., `TestMerge_ConflictingLoggers`)
- No globals (except slog.Logger in CLI entry point)

### Known Limitations (Out of v1 Scope)
- No recursive directory watching (config files must be in single watched directory)
- Single target type per config file (future: support multiple outputs per file)
- No TOML/JSON input formats (YAML only)
- No XSD validation for Logback output
- No built-in plugins/dynamic loading (registry is compile-time only)

## Critical Files Reference

| File | Purpose |
|------|---------|
| `cmd/outwatch/main.go` | CLI entry, flag parsing, registry setup, watch loop |
| `internal/config/schema.go` | FileConfig struct, validation rules, target-specific sections |
| `internal/config/parser.go` | YAML→ParsedFile, strict decoding, level normalization |
| `internal/discovery/fileset.go` | Directory scan for .yaml/.yml files |
| `internal/grouping/output_groups.go` | Group by (target, filename), deterministic sort |
| `internal/merge/merge.go` | Collision detection, first-file-wins fallback, warnings |
| `internal/pipeline/pipeline.go` | Central orchestrator, state management, reconcile() logic |
| `internal/targets/target.go` | Renderer interface definition |
| `internal/targets/registry.go` | Plugin registry (Register, Get) |
| `internal/targets/logback/render.go` | Logback XML renderer implementation |
| `internal/watch/watcher.go` | fsnotify wrapper with debounce + batching |
| `internal/output/paths.go` | Basename validation, path safety |
| `internal/output/writer.go` | Atomic write (temp + rename) |
| `.github/workflows/ci.yml` | Test automation on push/PR |
| `.github/workflows/release.yml` | Tag-triggered multi-platform build + image publish |
| `Dockerfile` | Multi-stage build, distroless nonroot runtime |
| `Makefile` | Build automation |
| `README.md` | Usage guide, config examples, deployment patterns |

## Quick Debugging Tips

**Config not parsing:**
- Check YAML syntax (indentation, quotes)
- Check `--log-level debug` output for validation error messages
- Inspect with `go run ./cmd/outwatch --once --input-dir ./configs --log-level debug`

**Outputs not generated:**
- Check output directory permissions (must be writable)
- Check log output for merge conflicts or invalid config errors
- Verify target name matches registered renderer (case-sensitive)

**Watcher not triggering:**
- Check fsnotify supports your filesystem (some network filesystems don't)
- Verify file operations complete before polling (some editors have async write buffers)
- Increase `--debounce` if events batch poorly

**Docker image build fails:**
- Run `make clean` to remove dist/ artifacts
- Check Dockerfile multi-stage syntax (builder → runtime)
- Verify distroless image is available: `docker pull gcr.io/distroless/static-debian12:nonroot`

## Contributing New Features

1. **Open issue** to discuss feature (collision merge strategy, extensibility impact, backward compat)
2. **Create feature branch:** `git checkout -b feat/description`
3. **Implement + test:** Ensure all tests pass, add tests for new behavior
4. **Update docs:** README section, GoDoc comments, .github/copilot-instructions.md if architecture changed
5. **Commit with conventional message:** `feat: add xyz functionality`
6. **Push and open PR** for review
7. **After merge:** Tag release and push to trigger automation

---

**Last Updated:** May 10, 2026  
**Current Version:** v0.2.0  
**Go Version:** 1.26  
**License:** See LICENSE file
