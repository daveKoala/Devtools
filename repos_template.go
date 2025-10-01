package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed template.yml
var embeddedTemplate []byte

// repoTemplate represents the shape of template.yml.
type repoTemplate struct {
	Services map[string]repoService `yaml:"services"`
}

// repoService captures the commands and relationships for a single service.
type repoService struct {
	Clone         string   `yaml:"clone"`
	PostCloneCmds []string `yaml:"postCloneCmds"`
	Depends       []string `yaml:"depends"`
}

// loadRepoTemplate fetches and parses template.yml, optionally falling back to the embedded copy.
func loadRepoTemplate(path string, allowEmbedded bool) (*repoTemplate, error) {
	contents, err := os.ReadFile(path)
	switch {
	case err == nil:
	case allowEmbedded && errors.Is(err, os.ErrNotExist):
		if len(embeddedTemplate) == 0 {
			return nil, fmt.Errorf("template %q not found and no embedded default available", path)
		}
		fmt.Printf("template %q not found, using embedded default\n", path)
		contents = embeddedTemplate
	default:
		return nil, fmt.Errorf("read template %q: %w", path, err)
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

// cloneOrder returns a dependency-safe ordering for all services in the template.
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

// cloneListFor returns the ordered set of services required for a single target service.
func (t *repoTemplate) cloneListFor(name string) ([]string, error) {
	if _, ok := t.Services[name]; !ok {
		return nil, fmt.Errorf("service %q not defined", name)
	}

	required := make(map[string]struct{})

	var mark func(string) error
	mark = func(svcName string) error {
		svc, ok := t.Services[svcName]
		if !ok {
			return fmt.Errorf("service %q not defined", svcName)
		}
		for _, dep := range svc.Depends {
			dep = strings.TrimSpace(dep)
			if dep == "" {
				continue
			}
			if _, seen := required[dep]; seen {
				continue
			}
			required[dep] = struct{}{}
			if err := mark(dep); err != nil {
				return err
			}
		}
		return nil
	}

	if err := mark(name); err != nil {
		return nil, err
	}

	required[name] = struct{}{}

	order, err := t.cloneOrder()
	if err != nil {
		return nil, err
	}

	sequence := make([]string, 0, len(required))
	for _, svcName := range order {
		if _, ok := required[svcName]; ok {
			sequence = append(sequence, svcName)
		}
	}

	return sequence, nil
}
