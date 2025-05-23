package docs

import (
	"fmt"
	"os"
	"path/filepath"
)

func renderGeneratedSection(packageRoot string, sectionName string) (string, error) {
	parentDir := filepath.Join(packageRoot, "_dev", "build", "docs", "sections")
	sectionFilename := fmt.Sprintf("generated_%s.md", sectionName)
	sectionPath := filepath.Join(parentDir, sectionFilename)
	body, err := os.ReadFile(sectionPath)
	if err != nil {
		return "", fmt.Errorf("reading section file failed (path: %s): %w", sectionPath, err)
	}

	return string(body), nil
}
