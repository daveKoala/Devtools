package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create context that handles graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Create task registry and register available tasks
	registry := NewTaskRegistry()

	// Register example tasks
	registry.Register(&HelloWorldTask{})
	registry.Register(&SystemInfoTask{})
	registry.Register(&BuildTask{})
	registry.Register(&TestTask{})
	registry.Register(&DependancyCheckTask{})

	// Create and display menu
	menu := NewMenu(registry)
	if err := menu.Display(ctx); err != nil {
		log.Fatalf("Menu error: %v", err)
	}
}
