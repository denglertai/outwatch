// Package grouping groups parsed config outputs for merge and render processing.
package grouping

import (
	"fmt"
	"sort"

	"github.com/denglertai/outwatch/internal/config"
)

// OutputGroup represents a set of source outputs targeting the same target/file pair.
type OutputGroup struct {
	Target string
	File   string
	Items  []config.ParsedOutput
}

// Build groups parsed files by target and output file and sorts deterministically.
func Build(files map[string]config.ParsedFile) []OutputGroup {
	grouped := map[string][]config.ParsedOutput{}
	for _, file := range files {
		output := file.Output
		key := groupKey(output.Config.Target, output.Config.File)
		grouped[key] = append(grouped[key], output)
	}

	keys := make([]string, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	groups := make([]OutputGroup, 0, len(keys))
	for _, key := range keys {
		items := grouped[key]
		sort.Slice(items, func(i, j int) bool {
			return items[i].SourcePath < items[j].SourcePath
		})
		target := items[0].Config.Target
		file := items[0].Config.File
		groups = append(groups, OutputGroup{Target: target, File: file, Items: items})
	}

	return groups
}

// groupKey builds a stable map key for a target/file pair.
func groupKey(target, file string) string {
	return fmt.Sprintf("%s|%s", target, file)
}
