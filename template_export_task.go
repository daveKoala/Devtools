package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// TemplateExportTask writes the embedded template to disk for developers to customise.
type TemplateExportTask struct {
	Destination string
}

func (t *TemplateExportTask) Name() string {
	return "Export Template"
}

func (t *TemplateExportTask) Description() string {
	return "Write embedded template.yml to exported_template.yml"
}

func (t *TemplateExportTask) Run(ctx context.Context) error {
	if len(embeddedTemplate) == 0 {
		return errors.New("no embedded template available")
	}

	dest := t.destinationPath()
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("ensure destination directory: %w", err)
	}

	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("%s already exists; remove or choose a different destination", dest)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check destination: %w", err)
	}

	if err := os.WriteFile(dest, embeddedTemplate, 0o644); err != nil {
		return fmt.Errorf("write template: %w", err)
	}

	fmt.Printf("Embedded template written to %s\n", dest)
	return nil
}

func (t *TemplateExportTask) destinationPath() string {
	if t.Destination != "" {
		return t.Destination
	}
	return "exported_template.yml"
}
