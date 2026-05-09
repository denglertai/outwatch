// Package merge tests merge and collision fallback behavior.
package merge

import (
	"testing"

	"github.com/denglertai/outwatch/internal/config"
	"github.com/denglertai/outwatch/internal/grouping"
)

// TestGroup_MergeCompatible verifies compatible logger maps are merged.
func TestGroup_MergeCompatible(t *testing.T) {
	group := grouping.OutputGroup{
		Target: "logback",
		File:   "app.xml",
		Items: []config.ParsedOutput{
			{SourcePath: "a.yaml", Config: config.OutputConfig{Target: "logback", File: "app.xml", Loggers: map[string]string{"a": "INFO"}}},
			{SourcePath: "b.yaml", Config: config.OutputConfig{Target: "logback", File: "app.xml", Loggers: map[string]string{"b": "DEBUG"}}},
		},
	}

	merged, warnings := Group(group)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if merged.Loggers["a"] != "INFO" || merged.Loggers["b"] != "DEBUG" {
		t.Fatalf("unexpected merge result: %#v", merged.Loggers)
	}
}

// TestGroup_ConflictKeepsFirst verifies conflict fallback keeps the first source.
func TestGroup_ConflictKeepsFirst(t *testing.T) {
	group := grouping.OutputGroup{
		Target: "logback",
		File:   "app.xml",
		Items: []config.ParsedOutput{
			{SourcePath: "a.yaml", Config: config.OutputConfig{Target: "logback", File: "app.xml", Loggers: map[string]string{"a": "INFO"}}},
			{SourcePath: "b.yaml", Config: config.OutputConfig{Target: "logback", File: "app.xml", Loggers: map[string]string{"a": "DEBUG"}}},
		},
	}

	merged, warnings := Group(group)
	if len(warnings) == 0 {
		t.Fatalf("expected warning")
	}
	if merged.Loggers["a"] != "INFO" {
		t.Fatalf("expected first file logger level, got %s", merged.Loggers["a"])
	}
}
