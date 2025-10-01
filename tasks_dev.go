package main

import (
	"context"
	"fmt"
)

// HelloWorldTask demonstrates the basic task interface
type HelloWorldTask struct{}

func (h *HelloWorldTask) Name() string {
	return "Hello world"
}

func (h *HelloWorldTask) Description() string {
	return "Overview of DevTools"
}

func (h *HelloWorldTask) Run(ctx context.Context) error {
	fmt.Println("Hello from the devtools!")
	fmt.Println()
	fmt.Println("This tool is meant to get new engineers productive quickly. Each menu item wraps one of our day-to-day workflows—cloning stacks, checking dependencies, running builds—so you can land in a working environment without memorising all the commands.")
	fmt.Println()
	fmt.Println("Start with the Clone Repos task to scaffold the services you need. It reads the shared template, pulls the repos, applies environment defaults, and even runs post-clone health checks so you know the containers are alive before you dive in.")
	fmt.Println()
	fmt.Println("As you get comfortable, explore the other entries: run system checks, build or test locally, and list your SSH keys when you need to register a new machine. Treat the menu as a living cookbook—add tasks when you find yourself repeating a command sequence, and everyone on the team benefits.")
	return nil
}
