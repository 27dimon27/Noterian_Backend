package queries

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
)

func LoadQueries(fs embed.FS, dir string) (map[string]string, error) {
	queries := make(map[string]string)

	files, err := fs.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read queries directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		content, err := fs.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read query file %s: %w", file.Name(), err)
		}

		key := strings.TrimSuffix(file.Name(), ".sql")
		queries[key] = strings.TrimSpace(string(content))
	}

	return queries, nil
}
