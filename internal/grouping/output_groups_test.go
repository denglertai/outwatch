// Package grouping tests output grouping behavior.
package grouping

import (
	"testing"

	"github.com/denglertai/outwatch/internal/config"
)

// TestBuild_GroupsAndSorts verifies grouping and deterministic ordering.
func TestBuild_GroupsAndSorts(t *testing.T) {
	files := map[string]config.ParsedFile{
		"b.yaml": {Path: "b.yaml", Output: config.ParsedOutput{SourcePath: "b.yaml", Config: config.OutputConfig{Target: "logback", File: "x.xml", Loggers: map[string]string{"b": "INFO"}}}},
		"a.yaml": {Path: "a.yaml", Output: config.ParsedOutput{SourcePath: "a.yaml", Config: config.OutputConfig{Target: "logback", File: "x.xml", Loggers: map[string]string{"a": "INFO"}}}},
		"c.yaml": {Path: "c.yaml", Output: config.ParsedOutput{SourcePath: "c.yaml", Config: config.OutputConfig{Target: "logback", File: "y.xml", Loggers: map[string]string{"c": "INFO"}}}},
	}

	groups := Build(files)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	if groups[0].File != "x.xml" || groups[1].File != "y.xml" {
		t.Fatalf("unexpected group order: %#v", groups)
	}
	if groups[0].Items[0].SourcePath != "a.yaml" || groups[0].Items[1].SourcePath != "b.yaml" {
		t.Fatalf("unexpected file order: %#v", groups[0].Items)
	}
}
