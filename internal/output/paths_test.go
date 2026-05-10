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

// TestResolvePath_Empty_FileName verifies empty filename is rejected.
func TestResolvePath_Empty_FileName(t *testing.T) {
	_, err := ResolvePath("/tmp/out", "")
	if err == nil {
		t.Fatal("expected error for empty filename")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Fatalf("expected empty error, got %v", err)
	}
}

// TestResolvePath_Whitespace_Only_FileName verifies whitespace-only filename is rejected.
func TestResolvePath_Whitespace_Only_FileName(t *testing.T) {
	_, err := ResolvePath("/tmp/out", "   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only filename")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Fatalf("expected empty error, got %v", err)
	}
}

// TestResolvePath_Backslash_Separator verifies backslash is also rejected.
func TestResolvePath_Backslash_Separator(t *testing.T) {
	_, err := ResolvePath("/tmp/out", "foo\\bar.xml")
	if err == nil {
		t.Fatal("expected error for backslash separator")
	}
	if !strings.Contains(err.Error(), "must not contain path") {
		t.Fatalf("expected path error, got %v", err)
	}
}
