package output

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWrite_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	payload := []byte("test content")

	err := AtomicWrite(filePath, payload)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("expected 'test content', got %q", string(content))
	}
}

func TestAtomicWrite_Overwrites_Existing(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Write initial content
	if err := os.WriteFile(filePath, []byte("old"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Overwrite with AtomicWrite
	newPayload := []byte("new content")
	err := AtomicWrite(filePath, newPayload)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(content) != "new content" {
		t.Errorf("expected 'new content', got %q", string(content))
	}
}

func TestAtomicWrite_Large_Payload(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Create a large payload
	payload := make([]byte, 1024*1024) // 1MB
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	err := AtomicWrite(filePath, payload)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if len(content) != len(payload) {
		t.Errorf("expected %d bytes, got %d", len(payload), len(content))
	}
}

func TestAtomicWrite_Invalid_Directory(t *testing.T) {
	// Use a non-existent directory
	filePath := "/nonexistent/directory/test.txt"

	err := AtomicWrite(filePath, []byte("test"))
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

func TestRemoveIfExists_File_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Create a file
	if err := os.WriteFile(filePath, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify it exists
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("File not created: %v", err)
	}

	// Remove it
	err := RemoveIfExists(filePath)
	if err != nil {
		t.Fatalf("RemoveIfExists failed: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(filePath); err == nil {
		t.Fatal("file should have been removed")
	}
}

func TestRemoveIfExists_File_Not_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.txt")

	// Should not error even though file doesn't exist
	err := RemoveIfExists(filePath)
	if err != nil {
		t.Fatalf("RemoveIfExists should not error for non-existent file: %v", err)
	}
}

func TestRemoveIfExists_Permission_Denied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("skipping permission test as root")
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Create a file
	if err := os.WriteFile(filePath, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Remove write permission from directory
	if err := os.Chmod(tmpDir, 0o555); err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}
	defer os.Chmod(tmpDir, 0o755) // restore for cleanup

	// Try to remove - should error
	err := RemoveIfExists(filePath)
	if err == nil {
		t.Fatal("expected error when removing file from read-only directory")
	}
}
