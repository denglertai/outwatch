# outwatch

[![CI](https://github.com/denglertai/outwatch/actions/workflows/ci.yml/badge.svg)](https://github.com/denglertai/outwatch/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/denglertai/outwatch/graph/badge.svg)](https://codecov.io/gh/denglertai/outwatch)

Outwatch watches a directory of YAML configuration files and generates structured outputs through an extensible target architecture.

V1 target: Logback XML.

## Features

- Watches a single non-recursive input directory.
- Supports both `.yaml` and `.yml` input files.
- Each config file declares exactly one output.
- Output definition uses top-level `target` and `filename`.
- Target-specific options are placed under a section matching the target name (for example `logback`).
- Output directory is configured through CLI (`--output-dir`).
- Multiple files targeting the same output file:
	- Merge when logger definitions are compatible.
	- If not mergeable, keep the first file in deterministic filename order and log a warning.
- Invalid config updates do not overwrite last valid generated output.
- Extensible output target registry (Logback implemented, more targets can be added).

## Input Config Format

```yaml
target: logback
filename: application-logback.xml
logback:
  loggers:
    com.example.service: INFO
    org.hibernate.SQL: DEBUG
```

Rules:

- `target` is required.
- `filename` is required and must be a file name only.
- For `target: logback`, `logback.loggers` must contain at least one logger.
- Allowed levels: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `OFF`.

## CLI

```bash
go run ./cmd/outwatch \
	--input-dir ./configs \
	--output-dir ./generated \
	--create-output-dir
```

Flags:

- `--input-dir` required.
- `--output-dir` required.
- `--debounce` event debounce duration (default `300ms`).
- `--log-level` log verbosity: `debug`, `info`, `warn`, `error` (default `info`).
- `--once` run one sync and exit.
- `--version` print build version metadata and exit.
- `--create-output-dir` create output directory when missing.
- `--delete-stale-outputs` remove generated files that no longer have source configs (default `true`).

Example version output:

```bash
outwatch --version
# outwatch version=v0.2.0 commit=abc123... buildDate=2026-05-10T12:00:00Z
```

## Complete Example Config

Example input file (`configs/log-levels.yaml`):

```yaml
target: logback
filename: dynamic-loggers.xml
logback:
  loggers:
    com.example.service: INFO
    com.example.api: DEBUG
    com.example.repository: WARN
    org.hibernate.SQL: DEBUG
    org.springframework.web: INFO
```

Run outwatch in watch mode:

```bash
go run ./cmd/outwatch \
	--input-dir ./configs \
	--output-dir ./generated \
	--create-output-dir \
	--log-level info
```

This continuously regenerates `generated/dynamic-loggers.xml` whenever files in `configs/` change.

## Logback Integration (With Watch)

Outwatch handles file watching for your YAML source files. Logback can also watch and reload `logback.xml` itself with `scan="true"`.

Example `logback.xml`:

```xml
<configuration scan="true" scanPeriod="10 seconds">
	<!-- Your normal appenders/root logger setup -->
	<appender name="STDOUT" class="ch.qos.logback.core.ConsoleAppender">
		<encoder>
			<pattern>%d %-5level [%thread] %logger - %msg%n</pattern>
		</encoder>
	</appender>

	<root level="INFO">
		<appender-ref ref="STDOUT" />
	</root>

	<!-- Load dynamically generated logger-level overrides -->
	<include optional="true" file="/opt/myapp/generated/dynamic-loggers.xml" />
</configuration>
```

Recommended runtime flow:

1. Start outwatch as a sidecar/process watching your config directory.
2. Outwatch writes updates into your configured output directory.
3. Logback applies generated logger changes via the include.

Note:

- Current logback target output is generated as a complete XML document containing logger entries.
- If your runtime requires include fragments with a different root element, align your Logback include strategy accordingly (or use the generated file as the primary Logback configuration file).

## Tests

Run all tests:

```bash
make test
```

## Deployment Examples

- Docker example (compose + local image build): `examples/docker/`
- Kubernetes example (sidecar pattern): `examples/kubernetes/`

Quick links:

- [examples/docker/README.md](./examples/docker/README.md)
- [examples/kubernetes/README.md](./examples/kubernetes/README.md)

## Container Image

Build the project image from the repository root:

```bash
make image-build
```

Build without cache:

```bash
make image-build-no-cache
```

Run the container:

```bash
docker run --rm \
	-v $(pwd)/configs:/config:ro \
	-v $(pwd)/generated:/generated \
	outwatch:latest \
	--input-dir /config \
	--output-dir /generated \
	--create-output-dir \
	--log-level info
```

The runtime image uses distroless nonroot (`gcr.io/distroless/static-debian12:nonroot`).

## Production Notes

- Run outwatch as a sidecar (Kubernetes) or sibling service (Docker Compose) with shared volume access to generated output.
- Ensure the output directory is writable by the outwatch runtime user:
	- **Kubernetes:** Use `securityContext.fsGroup: 0` on the Pod to make volumes group-writable without running as root.
	- **Docker Compose:** Use bind mounts with proper host directory permissions (e.g., `mkdir -p ./generated && chmod 777 ./generated`). See [examples/docker/README.md](./examples/docker/README.md).
- Invalid config updates are logged and ignored; the last valid generated output stays in place.
- Start order should guarantee outwatch can access mounted config and output volumes before first sync.
- Keep Logback include paths stable and mount generated output consistently across restarts.

## Makefile

Project automation targets are available in `Makefile`:

```bash
make help
make test
make build
make clean
make image-build
make image-build-no-cache
```

**Local testing targets:**

```bash
make test-docker-setup      # Create ./examples/docker/generated with correct permissions
make test-docker-up         # Start Docker Compose example
make test-docker-down       # Stop Docker Compose example
```

Common overrides:

```bash
make image-build IMAGE=ghcr.io/your-org/outwatch TAG=v0.2.0
make build BINARY=outwatch-linux-amd64 DIST_DIR=./build
```

Makefile parameters:

- `BINARY` (default: `outwatch`): output binary name for `make build`.
- `CMD_PATH` (default: `./cmd/outwatch`): Go package path used as the build entrypoint.
- `DIST_DIR` (default: `./dist`): output directory for build artifacts.
- `IMAGE` (default: `outwatch`): container image repository/name used by image build targets.
- `TAG` (default: `latest`): container image tag used by image build targets.

Parameterized examples:

```bash
make build CMD_PATH=./cmd/outwatch BINARY=outwatch-linux-arm64 DIST_DIR=./build
make image-build IMAGE=ghcr.io/your-org/outwatch TAG=v0.2.0
```

## CI and Dependency Updates

- CI runs tests on every push and pull request: `.github/workflows/ci.yml`.
- CI also runs `go test -race ./...` and `govulncheck ./...` for additional runtime and dependency safety checks.
- Renovate config is included in `renovate.json`:
	- Weekly schedule.
	- Grouped non-major updates for Go modules and GitHub Actions.
	- Automerge disabled.

## Release Automation and Branch Protection

- Tag push triggers `.github/workflows/release.yml` to:
	- Run tests.
	- Create a GitHub Release automatically.
	- Upload CLI assets (linux/darwin as `.tar.gz`, windows as `.zip`) and `checksums.txt`.
	- Build and publish image tags to GHCR (`<tag>` and `latest`).
- Recommended branch protection on default branch:
	- Require pull request before merge.
	- Require status checks to pass (`CI / test`).
	- Restrict direct pushes to protected branch.

## Extending with New Targets

1. Implement `internal/targets.Renderer`.
2. Register it in the CLI setup (currently in `cmd/outwatch/main.go`).
3. Run existing tests and add renderer-specific tests.
