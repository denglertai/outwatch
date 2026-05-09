// Package config contains configuration parsing and validation primitives.
package config

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

var validLevels = map[string]struct{}{
	"TRACE": {},
	"DEBUG": {},
	"INFO":  {},
	"WARN":  {},
	"ERROR": {},
	"OFF":   {},
}

// FileConfig is the YAML-backed root configuration model for one source file.
type FileConfig struct {
	Target   string         `yaml:"target"`
	FileName string         `yaml:"filename"`
	Logback  *LogbackConfig `yaml:"logback"`
	Extra    map[string]any `yaml:",inline"`
}

// ParsedFile represents a validated file with its normalized output definition.
type ParsedFile struct {
	Path   string
	Output ParsedOutput
}

// LogbackConfig contains logback-specific configuration fields.
type LogbackConfig struct {
	Loggers map[string]string `yaml:"loggers"`
}

// OutputConfig is the normalized, target-agnostic output definition used by the pipeline.
type OutputConfig struct {
	Target  string            `yaml:"target"`
	File    string            `yaml:"file"`
	Loggers map[string]string `yaml:"loggers"`
}

// ParsedOutput captures the source file and normalized output payload.
type ParsedOutput struct {
	SourcePath string
	Config     OutputConfig
}

// Validate checks semantic constraints and normalizes accepted values.
func (c *FileConfig) Validate() error {
	if strings.TrimSpace(c.Target) == "" {
		return fmt.Errorf("target is required")
	}
	c.Target = strings.ToLower(strings.TrimSpace(c.Target))

	if strings.TrimSpace(c.FileName) == "" {
		return fmt.Errorf("filename is required")
	}

	if filepath.Base(c.FileName) != c.FileName || strings.Contains(c.FileName, "/") || strings.Contains(c.FileName, "\\") {
		return fmt.Errorf("filename must be a file name only, without path segments")
	}

	if strings.TrimSpace(c.FileName) == "." || strings.TrimSpace(c.FileName) == ".." {
		return fmt.Errorf("filename must be a valid file name")
	}

	switch c.Target {
	case "logback":
		if c.Logback == nil {
			return fmt.Errorf("logback section is required when target is %q", c.Target)
		}
		if len(c.Logback.Loggers) == 0 {
			return fmt.Errorf("logback.loggers must contain at least one entry")
		}
		for loggerName, level := range c.Logback.Loggers {
			if strings.TrimSpace(loggerName) == "" {
				return fmt.Errorf("logger name must not be empty")
			}
			normalized := strings.ToUpper(strings.TrimSpace(level))
			if _, ok := validLevels[normalized]; !ok {
				return fmt.Errorf("invalid level %q for logger %q", level, loggerName)
			}
			c.Logback.Loggers[loggerName] = normalized
		}
	default:
		if _, ok := c.Extra[c.Target]; !ok {
			return fmt.Errorf("section %q is required when target is %q", c.Target, c.Target)
		}
	}

	return nil
}

// SortedLoggerNames returns logger names in deterministic lexical order.
func SortedLoggerNames(loggers map[string]string) []string {
	names := make([]string, 0, len(loggers))
	for name := range loggers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
