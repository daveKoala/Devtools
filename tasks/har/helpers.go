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

func readLine(ctx context.Context, scanner *bufio.Scanner) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if scanner.Scan() {
		return scanner.Text(), nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}

	return "", errors.New("input stream closed")
}

func expandUserPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	if path[0] != '~' {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}

	if path == "~" {
		return home, nil
	}

	if len(path) > 1 && (path[1] == '/' || path[1] == os.PathSeparator) {
		return filepath.Join(home, path[2:]), nil
	}

	return "", fmt.Errorf("unsupported user expansion in path %q", path)
}

func parseDomains(input string) []string {
	if input == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	cleaned := make([]string, 0, len(parts))
	seen := make(map[string]struct{})

	for _, part := range parts {
		domain := strings.ToLower(strings.TrimSpace(part))
		if domain == "" {
			continue
		}
		if _, exists := seen[domain]; exists {
			continue
		}
		seen[domain] = struct{}{}
		cleaned = append(cleaned, domain)
	}

	return cleaned
}
