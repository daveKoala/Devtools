package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SSHKeyTask lists SSH public keys so they can be copied into Bitbucket.
type SSHKeyTask struct {
	SearchDir string
}

func (t *SSHKeyTask) Name() string {
	return "List SSH Keys"
}

func (t *SSHKeyTask) Description() string {
	return "Show copy-ready ~/.ssh/*.pub entries"
}

func (t *SSHKeyTask) Run(ctx context.Context) error {
	searchDir, err := t.resolveSearchDir()
	if err != nil {
		return err
	}

	matches, err := filepath.Glob(filepath.Join(searchDir, "*.pub"))
	if err != nil {
		return fmt.Errorf("locate public keys: %w", err)
	}

	if len(matches) == 0 {
		fmt.Printf("No SSH public keys found in %s\n", searchDir)
		fmt.Println("Generate one with: ssh-keygen -t ed25519 -C \"you@example.com\"")
		return nil
	}

	fmt.Printf("Found %d SSH public key(s) in %s\n\n", len(matches), searchDir)

	for i, path := range matches {
		contents, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("%d. %s (unable to read: %v)\n\n", i+1, path, err)
			continue
		}

		firstLine := firstLine(strings.TrimSpace(string(contents)))
		keyType, comment := parseKeyMetadata(firstLine)

		fmt.Printf("%d. %s\n", i+1, path)
		fmt.Printf("   Type   : %s\n", keyType)
		if comment != "" {
			fmt.Printf("   Comment: %s\n", comment)
		}
		fmt.Println("   --- Copy below ---")
		fmt.Println(firstLine)
		fmt.Println("   ------------------")
		fmt.Println()
	}

	return nil
}

func (t *SSHKeyTask) resolveSearchDir() (string, error) {
	if t.SearchDir != "" {
		return t.SearchDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	return filepath.Join(home, ".ssh"), nil
}

func firstLine(text string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func parseKeyMetadata(line string) (string, string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "unknown", ""
	}
	keyType := fields[0]

	comment := ""
	if len(fields) >= 3 {
		comment = strings.Join(fields[2:], " ")
	} else if len(fields) == 2 {
		comment = fields[1]
	}

	return keyType, comment
}
