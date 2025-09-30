package main

import "context"

// Task defines the interface that all devtools must implement
type Task interface {
	// Name returns the display name for the menu
	Name() string

	// Description returns a brief description of what the task does
	Description() string

	// Run executes the task with the given context
	Run(ctx context.Context) error
}

// TaskRegistry manages available tasks
type TaskRegistry struct {
	tasks []Task
}

// NewTaskRegistry creates a new task registry
func NewTaskRegistry() *TaskRegistry {
	return &TaskRegistry{
		tasks: make([]Task, 0),
	}
}

// Register adds a task to the registry
func (tr *TaskRegistry) Register(task Task) {
	tr.tasks = append(tr.tasks, task)
}

// GetTasks returns all registered tasks
func (tr *TaskRegistry) GetTasks() []Task {
	return tr.tasks
}

// GetTask returns a task by index (0-based)
func (tr *TaskRegistry) GetTask(index int) Task {
	if index < 0 || index >= len(tr.tasks) {
		return nil
	}
	return tr.tasks[index]
}
