package main

import (
	"context"
	"fmt"
)

// HelloWorldTask demonstrates the basic task interface
type HelloWorldTask struct{}

func (h *HelloWorldTask) Name() string {
	return "Hello World"
}

func (h *HelloWorldTask) Description() string {
	return "Print a greeting message"
}

func (h *HelloWorldTask) Run(ctx context.Context) error {
	fmt.Println("Hello from the devtools!")
	return nil
}
