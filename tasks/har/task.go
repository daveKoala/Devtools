package har

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Task struct {
	LastHARPath string
	LastDomains []string
}

func New() *Task {
	return &Task{}
}

func (t *Task) Name() string {
	return "HAR Helper"
}

func (t *Task) Description() string {
	return "Collect a HAR file path and domains to inspect"
}

func (t *Task) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)

	harPath, err := t.promptForHARPath(ctx, scanner)
	if err != nil {
		return err
	}

	domains, err := t.promptForDomains(ctx, scanner)
	if err != nil {
		return err
	}

	t.LastHARPath = harPath
	t.LastDomains = domains

	outputPath, spec, err := processHarFile(ctx, harPath, domains)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("HAR file  : %s\n", harPath)
	fmt.Printf("Domains   : %s\n", strings.Join(domains, ", "))
	fmt.Printf("OpenAPI   : %s\n", outputPath)
	fmt.Printf("Path count: %d\n", len(spec.Paths))

	return nil
}

func processHarFile(ctx context.Context, path string, domains []string) (string, *OpenAPIDocument, error) {
	operation := NewProcessor(domains)
	operation.AllowHeader("Authorization", "Bearer {{BearerAdmin}}")
	doc, err := LoadDocument(ctx, path)
	if err != nil {
		return "", nil, err
	}

	payload, spec, err := operation.GenerateOpenAPIDocument(ctx, doc)
	if err != nil {
		return "", nil, err
	}

	outputPath := defaultOpenAPIPath(path)
	if err := os.WriteFile(outputPath, payload, 0o644); err != nil {
		return "", nil, fmt.Errorf("write OpenAPI document: %w", err)
	}

	return outputPath, spec, nil
}

func defaultOpenAPIPath(harPath string) string {
	base := strings.TrimSuffix(harPath, filepath.Ext(harPath))
	if base == "" {
		return harPath + ".openapi.json"
	}
	return base + ".openapi.json"
}
