package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
		repoPath, alreadyExists, err := cloneService(ctx, targetDir, name, svc.Clone)
		if err != nil {
			return err
		}

		if len(svc.PostCloneCmds) == 0 || alreadyExists {
			continue
		}

		if err := runPostCloneCommands(ctx, repoPath, name, svc.PostCloneCmds, svc.Environment); err != nil {
			return err
		}
	}
	return nil
}

// cloneService executes a git clone command in the target directory, skipping work that already exists.
func cloneService(ctx context.Context, targetDir, serviceName, cloneCmd string) (string, bool, error) {
	fields := strings.Fields(cloneCmd)
	if len(fields) < 3 {
		return "", false, fmt.Errorf("service %q: clone command must look like 'git clone <repo> [dir]'", serviceName)
	}
	if fields[0] != "git" || fields[1] != "clone" {
		return "", false, fmt.Errorf("service %q: clone command must start with 'git clone'", serviceName)
	}

	args := fields[2:]
	if len(args) > 2 {
		return "", false, fmt.Errorf("service %q: clone command only supports one optional target directory", serviceName)
	}

	repoURL := args[0]
	explicitDir := ""
	if len(args) == 2 {
		explicitDir = args[1]
	}

	repoDir, err := deriveRepoDir(repoURL, explicitDir)
	if err != nil {
		return "", false, fmt.Errorf("service %q: %w", serviceName, err)
	}

	clonePath := filepath.Join(targetDir, repoDir)
	if _, err := os.Stat(clonePath); err == nil {
		fmt.Printf("[%s] already exists at %s, skipping clone\n", serviceName, clonePath)
		return clonePath, true, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, fmt.Errorf("service %q: unable to inspect %s: %w", serviceName, clonePath, err)
	}

	cmd := exec.CommandContext(ctx, fields[0], fields[1:]...)
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("[%s] cloning %s into %s\n", serviceName, repoURL, clonePath)
	if err := cmd.Run(); err != nil {
		return "", false, fmt.Errorf("service %q: clone failed: %w", serviceName, err)
	}

	return clonePath, false, nil
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

// runPostCloneCommands executes post-clone commands inside the freshly cloned repository.
func runPostCloneCommands(ctx context.Context, repoPath, serviceName string, commands []string, envDefaults map[string]string) error {
	env := mergedEnv(envDefaults)

	for _, raw := range commands {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		fmt.Printf("[%s] post-clone: %s\n", serviceName, raw)
		cmd := exec.CommandContext(ctx, "bash", "-lc", raw)
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = env

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("service %q: post-clone command failed (%s): %w", serviceName, raw, err)
		}
	}
	return nil
}

func mergedEnv(defaults map[string]string) []string {
	values := map[string]string{}
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		key := parts[0]
		val := ""
		if len(parts) == 2 {
			val = parts[1]
		}
		values[key] = val
	}

	for key, val := range defaults {
		if val == "" {
			continue
		}
		if existing, ok := values[key]; !ok || existing == "" {
			values[key] = val
		}
	}

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]string, 0, len(keys))
	for _, k := range keys {
		result = append(result, fmt.Sprintf("%s=%s", k, values[k]))
	}
	return result
}
