# Go DevTools

A simple, menu-driven CLI tool for common development tasks. Built following pragmatic development principles.

## Usage

```bash
go run .
```

This displays a numbered menu. Enter a number to execute that task, or the last number to exit.

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