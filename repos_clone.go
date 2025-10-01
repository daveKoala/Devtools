package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureTargetDir creates the clone destination if it does not already exist.
func ensureTargetDir(targetDir string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	return nil
}

// cloneServices iterates the requested services and clones each one in turn.
func cloneServices(ctx context.Context, targetDir string, template *repoTemplate, names []string) error {

	if err := ensureTargetDir(targetDir); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	for _, name := range names {
		svc := template.Services[name]
		if err := cloneService(ctx, targetDir, name, svc.Clone); err != nil {
			return err
		}
	}
	return nil
}

// cloneService executes a git clone command in the target directory, skipping work that already exists.
func cloneService(ctx context.Context, targetDir, serviceName, cloneCmd string) error {
	fields := strings.Fields(cloneCmd)
	if len(fields) < 3 {
		return fmt.Errorf("service %q: clone command must look like 'git clone <repo> [dir]'", serviceName)
	}
	if fields[0] != "git" || fields[1] != "clone" {
		return fmt.Errorf("service %q: clone command must start with 'git clone'", serviceName)
	}

	args := fields[2:]
	if len(args) > 2 {
		return fmt.Errorf("service %q: clone command only supports one optional target directory", serviceName)
	}

	repoURL := args[0]
	explicitDir := ""
	if len(args) == 2 {
		explicitDir = args[1]
	}

	repoDir, err := deriveRepoDir(repoURL, explicitDir)
	if err != nil {
		return fmt.Errorf("service %q: %w", serviceName, err)
	}

	clonePath := filepath.Join(targetDir, repoDir)
	if _, err := os.Stat(clonePath); err == nil {
		fmt.Printf("[%s] already exists at %s, skipping clone\n", serviceName, clonePath)
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("service %q: unable to inspect %s: %w", serviceName, clonePath, err)
	}

	cmd := exec.CommandContext(ctx, fields[0], fields[1:]...)
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("[%s] cloning %s into %s\n", serviceName, repoURL, clonePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("service %q: clone failed: %w", serviceName, err)
	}

	return nil
}

// deriveRepoDir determines the local directory name for a repository.
func deriveRepoDir(repoURL, explicitDir string) (string, error) {
	if explicitDir != "" {
		return explicitDir, nil
	}

	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return "", errors.New("empty repository URL")
	}

	name := filepath.Base(repoURL)
	name = strings.TrimSuffix(name, ".git")
	name = strings.TrimSuffix(name, string(filepath.Separator))

	if name == "" || name == "." || name == string(filepath.Separator) {
		return "", fmt.Errorf("unable to determine repository directory from %q", repoURL)
	}

	return name, nil
}
