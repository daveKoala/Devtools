package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// HelloWorldTask demonstrates the basic task interface
type HelloWorldTask struct{}

func (h *HelloWorldTask) Name() string {
	return "Hello World"
}

func (h *HelloWorldTask) Description() string {
	return "Print a greeting message"
}

func (h *HelloWorldTask) Run(ctx context.Context) error {
	fmt.Println("Hello from the devtools!")
	return nil
}

// BuildTask runs go build
type BuildTask struct{}

func (b *BuildTask) Name() string {
	return "Build Project"
}

func (b *BuildTask) Description() string {
	return "Run 'go build' to compile the project"
}

func (b *BuildTask) Run(ctx context.Context) error {
	fmt.Println("Building project...")

	cmd := exec.CommandContext(ctx, "go", "build", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("Build completed successfully!")
	return nil
}

// TestTask runs go test
type TestTask struct{}

func (t *TestTask) Name() string {
	return "Run Tests"
}

func (t *TestTask) Description() string {
	return "Run 'go test' to execute all tests"
}

func (t *TestTask) Run(ctx context.Context) error {
	fmt.Println("Running tests...")

	cmd := exec.CommandContext(ctx, "go", "test", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	fmt.Println("All tests passed!")
	return nil
}
