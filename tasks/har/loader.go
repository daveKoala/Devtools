package har

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// LoadDocument reads a HAR file from disk and unmarshals it into a Document.
func LoadDocument(ctx context.Context, path string) (*Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open HAR file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, 25<<20)) // guard against unexpectedly large files
	if err != nil {
		return nil, fmt.Errorf("read HAR file: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse HAR JSON: %w", err)
	}

	return &doc, nil
}
