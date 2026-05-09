// Package config contains configuration parsing and validation primitives.
package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFile reads, decodes, validates, and normalizes a single config file.
func ParseFile(path string) (ParsedFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return ParsedFile{}, fmt.Errorf("read config %s: %w", path, err)
	}

	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(true)

	var cfg FileConfig
	if err := dec.Decode(&cfg); err != nil {
		return ParsedFile{}, fmt.Errorf("parse YAML %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return ParsedFile{}, fmt.Errorf("validate %s: %w", path, err)
	}

	outputCfg := OutputConfig{
		Target: cfg.Target,
		File:   cfg.FileName,
	}

	if cfg.Logback != nil {
		outputCfg.Loggers = make(map[string]string, len(cfg.Logback.Loggers))
		for k, v := range cfg.Logback.Loggers {
			outputCfg.Loggers[k] = v
		}
	}

	return ParsedFile{Path: path, Output: ParsedOutput{SourcePath: path, Config: outputCfg}}, nil
}
