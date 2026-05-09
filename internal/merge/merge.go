// Package merge resolves collisions between multiple inputs targeting one output.
package merge

import (
	"fmt"

	"github.com/denglertai/outwatch/internal/config"
	"github.com/denglertai/outwatch/internal/grouping"
)

// Group merges compatible output configs in a group or falls back to the first with warnings.
func Group(group grouping.OutputGroup) (config.OutputConfig, []string) {
	if len(group.Items) == 0 {
		return config.OutputConfig{Target: group.Target, File: group.File, Loggers: map[string]string{}}, nil
	}

	base := clone(group.Items[0].Config)
	warnings := make([]string, 0)

	for idx := 1; idx < len(group.Items); idx++ {
		candidate := group.Items[idx]
		if !mergeCompatible(base.Loggers, candidate.Config.Loggers) {
			warnings = append(warnings, fmt.Sprintf("non-mergeable config collision for output %q target %q, keeping first source %q and ignoring %q", group.File, group.Target, group.Items[0].SourcePath, candidate.SourcePath))
			return base, warnings
		}
		for logger, level := range candidate.Config.Loggers {
			base.Loggers[logger] = level
		}
	}

	return base, warnings
}

// mergeCompatible reports whether candidate loggers can be merged into base without conflicts.
func mergeCompatible(base map[string]string, candidate map[string]string) bool {
	for logger, level := range candidate {
		if existing, ok := base[logger]; ok && existing != level {
			return false
		}
	}
	return true
}

// clone makes a deep copy of an output config.
func clone(cfg config.OutputConfig) config.OutputConfig {
	out := config.OutputConfig{Target: cfg.Target, File: cfg.File, Loggers: make(map[string]string, len(cfg.Loggers))}
	for k, v := range cfg.Loggers {
		out.Loggers[k] = v
	}
	return out
}
