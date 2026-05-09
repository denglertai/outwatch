// Package output provides path resolution and file write helpers for generated outputs.
package output

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResolvePath validates an output file name and joins it with the output directory.
func ResolvePath(outputDir string, fileName string) (string, error) {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return "", fmt.Errorf("output file name must not be empty")
	}
	if filepath.Base(trimmed) != trimmed || strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\") {
		return "", fmt.Errorf("output file name %q must not contain path separators", fileName)
	}
	return filepath.Join(outputDir, trimmed), nil
}
