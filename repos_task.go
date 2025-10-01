package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultTemplatePath = "template.yml"
	defaultRepoDir      = "dev-app"
)

// ReposTask drives the interactive menu for cloning repositories.
type ReposTask struct {
	TemplatePath string
	TargetDir    string
}

// Name returns the menu label for this task.
func (s *ReposTask) Name() string {
	return "Clone Repos"
}

// Description returns a short explanation of what the task does.
func (s *ReposTask) Description() string {
	return "Clone repositories defined in template.yml"
}

// Run presents a submenu that lets developers clone every repo or a single service with its dependencies.
func (s *ReposTask) Run(ctx context.Context) error {
	templatePath := s.templatePath()
	template, err := loadRepoTemplate(templatePath, templatePath == defaultTemplatePath)
	if err != nil {
		return err
	}

	targetDir := s.targetDir()

	order, err := template.cloneOrder()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n=== Clone Repos ===")
		fmt.Println("1. Clone all services")

		for i, name := range order {
			svc := template.Services[name]
			deps := formatDependencies(svc.Depends)
			fmt.Printf("%d. %s (depends: %s)\n", i+2, name, deps)
		}

		backOption := len(order) + 2
		exitOption := len(order) + 3

		fmt.Printf("\n")
		fmt.Printf("%d. Back to main menu\n", backOption)
		fmt.Printf("%d. Exit\n", exitOption)
		fmt.Print("\nSelect option: ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			return errors.New("input stream closed")
		}

		choiceText := strings.TrimSpace(scanner.Text())
		choice, err := strconv.Atoi(choiceText)
		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}

		switch {
		case choice == 1:
			if err := cloneServices(ctx, targetDir, template, order); err != nil {
				return err
			}
		case choice >= 2 && choice < backOption:
			idx := choice - 2
			name := order[idx]
			cloneList, err := template.cloneListFor(name)
			if err != nil {
				return err
			}

			fmt.Printf("\nCloning sequence: %s\n", strings.Join(cloneList, ", "))
			if err := cloneServices(ctx, targetDir, template, cloneList); err != nil {
				return err
			}
		case choice == backOption:
			return nil
		case choice == exitOption:
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

// templatePath returns the configured template path or the default.
func (s *ReposTask) templatePath() string {
	if s.TemplatePath != "" {
		return s.TemplatePath
	}
	return defaultTemplatePath
}

// targetDir returns the directory that repositories should be cloned into.
func (s *ReposTask) targetDir() string {
	if s.TargetDir != "" {
		return s.TargetDir
	}
	return defaultRepoDir
}

// formatDependencies renders a user-friendly dependency list for menu output.
func formatDependencies(deps []string) string {
	cleaned := make([]string, 0, len(deps))
	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		cleaned = append(cleaned, dep)
	}
	if len(cleaned) == 0 {
		return "none"
	}
	return strings.Join(cleaned, ", ")
}
