package main

import (
	"context"
	"fmt"
	"os/exec"
)

type ReposTask struct{}

func (s *ReposTask) Name() string {
	return "repos"
}

func (s *ReposTask) Description() string {
	return "List all repos"
}

func (s *ReposTask) Run(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "gh", "repo", "list", "--limit", "1000")
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Println(string(stdout))

	return nil
}
