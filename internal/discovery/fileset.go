// Package discovery provides filesystem discovery utilities for config files.
package discovery

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// IsConfigFile reports whether a path has a supported YAML extension.
func IsConfigFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

// Discover lists and sorts supported config files in a non-recursive directory.
func Discover(inputDir string) ([]string, error) {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !IsConfigFile(name) {
			continue
		}
		paths = append(paths, filepath.Join(inputDir, name))
	}

	sort.Strings(paths)
	return paths, nil
}
