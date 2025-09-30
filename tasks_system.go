package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// SystemInfoTask shows system information
type SystemInfoTask struct{}

func (s *SystemInfoTask) Name() string {
	return "SystemInfo"

}

func (s *SystemInfoTask) Description() string {
	return "Shows system information"

}

func (s *SystemInfoTask) Run(ctx context.Context) error {
	fmt.Println("=== System Information ===")

	// Get current working directory
	if wd, err := os.Getwd(); err == nil {
		fmt.Printf("Working Directory: %s\n", wd)
	}

	// Get current time
	fmt.Printf("Current Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// Get Go version
	if output, err := exec.CommandContext(ctx, "go", "version").Output(); err == nil {
		fmt.Printf("Go Version: %s", string(output))
	}

	return nil
}

// DependencyCheckTask verifies required tools are installed
type DependancyCheckTask struct{}

func (d *DependancyCheckTask) Name() string {
	return "Check Dependencies"
}

func (d *DependancyCheckTask) Description() string {
	return "Verify that all required tools are installed"
}

func (d *DependancyCheckTask) Run(ctx context.Context) error {
	fmt.Println("Checking dependencies...")

	// Check git
	if err := checkCommand(ctx, "git", "--version"); err != nil {
		fmt.Printf("❌ git: %v\n", err)
	} else {
		fmt.Println("Git is installed")
	}

	// Check docker
	if err := checkCommand(ctx, "docker", "--version"); err != nil {
		fmt.Printf("❌ docker: %v\n", err)
	} else {
		fmt.Println("Docker is installed")
	}

	// Check docker-compose
	if err := checkCommand(ctx, "docker-compose", "--version"); err != nil {
		fmt.Printf("❌ docker-compose: %v\n", err)
	} else {
		fmt.Println("Docker Compose is installed")
	}

	return nil
}

func checkCommand(ctx context.Context, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	return cmd.Run()
}
