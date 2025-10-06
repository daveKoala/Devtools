package har

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (t *Task) promptForHARPath(ctx context.Context, scanner *bufio.Scanner) (string, error) {
	for {
		prompt := "Enter HAR file path: "
		if t.LastHARPath != "" {
			prompt = fmt.Sprintf("Enter HAR file path [%s]: ", t.LastHARPath)
		}
		fmt.Print(prompt)

		input, err := readLine(ctx, scanner)
		if err != nil {
			return "", err
		}

		raw := strings.TrimSpace(input)
		if raw == "" {
			if t.LastHARPath != "" {
				return t.LastHARPath, nil
			}
			fmt.Println("HAR file path cannot be empty.")
			continue
		}

		expanded, err := expandUserPath(raw)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}

		full := os.ExpandEnv(expanded)

		if _, err := os.Stat(full); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Println("File does not exist. Please try again.")
			} else {
				fmt.Printf("Unable to access file: %v\n", err)
			}
			continue
		}

		abs, err := filepath.Abs(full)
		if err != nil {
			fmt.Printf("Unable to resolve absolute path: %v\n", err)
			continue
		}

		return abs, nil
	}
}

func (t *Task) promptForDomains(ctx context.Context, scanner *bufio.Scanner) ([]string, error) {
	for {
		defaultHint := ""
		if len(t.LastDomains) > 0 {
			defaultHint = fmt.Sprintf(" [%s]", strings.Join(t.LastDomains, ", "))
		}
		fmt.Printf("Enter comma separated domains%s: ", defaultHint)

		input, err := readLine(ctx, scanner)
		if err != nil {
			return nil, err
		}

		raw := strings.TrimSpace(input)
		if raw == "" && len(t.LastDomains) > 0 {
			return t.LastDomains, nil
		}

		domains := parseDomains(raw)
		if len(domains) == 0 {
			fmt.Println("Please enter at least one domain.")
			continue
		}

		return domains, nil
	}
}
