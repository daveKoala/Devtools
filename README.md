# Go DevTools

A simple, menu-driven CLI tool for common development tasks. Built following pragmatic development principles.

## Usage

```bash
go run .
```

This displays a numbered menu. Enter a number to execute that task, or the last number to exit. The menu header shows the build version (derived from the Git tag/commit and build timestamp).

## Architecture

- **Task Interface**: All tools implement the `Task` interface with `Name()`, `Description()`, and `Run()` methods
- **TaskRegistry**: Manages and provides access to available tasks
- **Menu**: Handles user interaction and task execution

## Adding New Tasks

1. Create a struct that implements the `Task` interface:

```go
type MyTask struct{}

func (m *MyTask) Name() string {
    return "My Task"
}

func (m *MyTask) Description() string {
    return "Does something useful"
}

func (m *MyTask) Run(ctx context.Context) error {
    // Your task implementation
    return nil
}
```

2. Register it in `main.go`:

```go
registry.Register(&MyTask{})
```

## Included Tasks

- **Hello World**: Basic demonstration task
- **System Info**: Shows working directory, time, and Go version
- **Build Project**: Runs `go build`
- **Run Tests**: Runs `go test ./...`
- **Check Dependencies**: Verifies core tooling (git, docker, docker-compose)
- **Clone Repos**: Reads `template.yml` (embedded fallback) and clones repos with dependency ordering
- **Clone Repos** also applies any `environment` defaults before running post-clone commands
- **List SSH Keys**: Prints copy-ready SSH public keys for Bitbucket/GitHub setup
- **Export Template**: Writes the embedded `template.yml` to disk so teammates can customise their own copy

## Building & Packaging

Use the Makefile to build reproducible binaries with embedded version metadata:

```bash
# Run locally with git-derived version info
make run

# Build for your current platform (binary in bin/<os_arch>/)
make build

# Produce zip packages for Apple Silicon & Intel macOS
make package-macos

# Target a specific OS/arch manually (example: Windows)
make package TARGET_OS=windows TARGET_ARCH=amd64
```

`make build` injects version data using:

- `VERSION`: defaults to `git describe --tags --always`
- `COMMIT`: defaults to the short Git SHA
- `BUILD_DATE`: defaults to the UTC timestamp at build time

Override them when building outside Git (e.g. CI):

```bash
make package VERSION=v1.2.0 COMMIT=abc123 BUILD_DATE=2024-10-01T12:00:00Z
```

Each package contains a single `devtools` binary; unzip it and run `./devtools` on the target machine. macOS users may need to allow the binary under *System Settings → Privacy & Security* on first launch.

## Continuous Integration

GitHub Actions (`.github/workflows/ci.yml`) runs automatically on pull requests and pushes to `main`. The workflow:

- Checks formatting (`gofmt -l`)
- Runs `go test ./...` and `go build`
- When a tag matching `v*` is pushed, builds macOS (arm64 & amd64) and Linux AMD64 packages via the Makefile and uploads them as build artifacts

To cut a release:

```bash
git checkout main
git pull
git tag v1.0.0
git push origin v1.0.0
```

The CI job will attach the packaged zips to the workflow artifacts for that tag; you can publish them as a GitHub Release if desired.

## Template Overview

`template.yml` drives the repository cloning task. Each service entry supports:

- `clone`: full `git clone` command
- `depends`: list of services that must be cloned first
- `postCloneCmds`: shell commands executed in order after a fresh clone
- `environment`: key/value pairs exposed when `postCloneCmds` run (acting as defaults if the variable is not already set)
- `healthCheck`: optional command (with retries/intervals) executed after post-clone commands succeed

Example snippet:

```yaml
services:
  example:
    clone: git clone git@example.com/example.git
    depends:
      - core-api
    environment:
      API_PORT: "8081"
      DB_PASSWORD: supersecret
    healthCheck:
      command: "curl -f http://localhost:8081/health"
      interval: 5s
      retries: 6
      timeout: 3s
    postCloneCmds:
      - cp -n .env.example .env
      - docker compose up -d
```

When the service is cloned, the commands execute inside the repo directory with `API_PORT` and `DB_PASSWORD` available (unless already provided in the user’s shell). Existing clones are left untouched so local changes aren’t overwritten.

If a `healthCheck` block is provided, the tool runs the command using `bash -lc` after post-clone commands succeed. It retries up to `retries` times (default 5) with the specified `interval` (default 5s) and honours an optional per-attempt `timeout`. Environment defaults apply during the health check as well.

Need a copy you can tweak? Run the “Export Template” task and it will write the embedded YAML (with current defaults) to `exported_template.yml`. From there you can adjust paths or environments locally without changing the baked-in defaults.

Let me know what you think