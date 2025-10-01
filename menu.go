package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Menu handles the interactive menu system
type Menu struct {
	registry *TaskRegistry
	scanner  *bufio.Scanner
}

// NewMenu creates a new menu with the given task registry
func NewMenu(registry *TaskRegistry) *Menu {
	return &Menu{
		registry: registry,
		scanner:  bufio.NewScanner(os.Stdin),
	}
}

// Display shows the available options and handles user input
func (m *Menu) Display(ctx context.Context) error {
	for {
		clearTerminal()
		fmt.Printf("\n=== DevTools (version %s) ===\n", displayVersion())
		tasks := m.registry.GetTasks()

		if len(tasks) == 0 {
			fmt.Println("No tasks available.")
			return nil
		}

		for i, task := range tasks {
			fmt.Printf("%d. %s - %s\n", i+1, task.Name(), task.Description())
		}
		fmt.Printf("\n")
		fmt.Printf("%d. Exit\n", len(tasks)+1)

		fmt.Print("\nSelect option: ")

		if !m.scanner.Scan() {
			return fmt.Errorf("failed to read input")
		}

		input := strings.TrimSpace(m.scanner.Text())
		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}

		// Exit option
		if choice == len(tasks)+1 {
			fmt.Println("Goodbye!")
			return nil
		}

		// Execute selected task
		task := m.registry.GetTask(choice - 1)
		if task == nil {
			fmt.Println("Invalid option. Please try again.")
			continue
		}

		fmt.Printf("\nExecuting: %s\n", task.Name())
		if err := task.Run(ctx); err != nil {
			fmt.Printf("Error running task: %v\n", err)
		}

		fmt.Println("\nPress Enter to continue...")
		m.scanner.Scan()
	}
}

func clearTerminal() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	default:
		fmt.Print("\033[H\033[2J")
	}
}
