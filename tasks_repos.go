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

	"gopkg.in/yaml.v3"
)

const (
	defaultTemplatePath = "template.yml"
	defaultRepoDir      = "dev-app"
)

type ReposTask struct {
	TemplatePath string
	TargetDir    string
}

func (s *ReposTask) Name() string {
	return "Clone Repos"
}

func (s *ReposTask) Description() string {
	return "Clone repositories defined in template.yml"
}

func (s *ReposTask) Run(ctx context.Context) error {
	template, err := loadRepoTemplate(s.templatePath())
	if err != nil {
		return err
	}

	targetDir := s.targetDir()
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	order, err := template.cloneOrder()
	if err != nil {
		return err
	}

	for _, name := range order {
		svc := template.Services[name]
		if err := cloneService(ctx, targetDir, name, svc.Clone); err != nil {
			return err
		}
	}

	return nil
}

func (s *ReposTask) templatePath() string {
	if s.TemplatePath != "" {
		return s.TemplatePath
	}
	return defaultTemplatePath
}

func (s *ReposTask) targetDir() string {
	if s.TargetDir != "" {
		return s.TargetDir
	}
	return defaultRepoDir
}

type repoTemplate struct {
	Services map[string]repoService `yaml:"services"`
}

type repoService struct {
	Clone         string   `yaml:"clone"`
	PostCloneCmds []string `yaml:"postCloneCmds"`
	Depends       []string `yaml:"depends"`
}

func loadRepoTemplate(path string) (*repoTemplate, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}

	var tpl repoTemplate
	if err := yaml.Unmarshal(contents, &tpl); err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	if len(tpl.Services) == 0 {
		return nil, errors.New("template has no services defined")
	}

	return &tpl, nil
}

func (t *repoTemplate) cloneOrder() ([]string, error) {
	names := make([]string, 0, len(t.Services))
	for name := range t.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	visited := make(map[string]bool, len(t.Services))
	visiting := make(map[string]bool, len(t.Services))
	order := make([]string, 0, len(t.Services))

	var visit func(string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		if visiting[name] {
			return fmt.Errorf("detected circular dependency involving %q", name)
		}

		svc, ok := t.Services[name]
		if !ok {
			return fmt.Errorf("service %q not defined", name)
		}

		visiting[name] = true
		for _, dep := range svc.Depends {
			dep = strings.TrimSpace(dep)
			if dep == "" {
				continue
			}
			if _, ok := t.Services[dep]; !ok {
				return fmt.Errorf("service %q depends on unknown service %q", name, dep)
			}
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		order = append(order, name)
		return nil
	}

	for _, name := range names {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return order, nil
}

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
