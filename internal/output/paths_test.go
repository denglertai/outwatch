// Package output tests output path resolution helpers.
package output

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestResolvePath_Valid verifies valid output names are resolved correctly.
func TestResolvePath_Valid(t *testing.T) {
	path, err := ResolvePath("/tmp/out", "a.xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != filepath.Join("/tmp/out", "a.xml") {
		t.Fatalf("unexpected path: %s", path)
	}
}

// TestResolvePath_PathSeparatorRejected verifies path separators are rejected.
func TestResolvePath_PathSeparatorRejected(t *testing.T) {
	_, err := ResolvePath("/tmp/out", "foo/bar.xml")
	if err == nil || !strings.Contains(err.Error(), "must not contain path") {
		t.Fatalf("expected error, got %v", err)
	}
}
