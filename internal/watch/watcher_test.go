// Package watch tests filesystem watch orchestration behavior.
package watch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDirectory_EmitsBatchedEvents verifies create and remove events are emitted.
func TestDirectory_EmitsBatchedEvents(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan []string, 8)
	errCh := make(chan error, 1)

	go func() {
		errCh <- Directory(ctx, dir, 50*time.Millisecond, func(paths []string) {
			cp := make([]string, len(paths))
			copy(cp, paths)
			events <- cp
		})
	}()

	// Allow watcher setup to complete before triggering fs events.
	time.Sleep(150 * time.Millisecond)

	cfgPath := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte("target: logback\nfilename: x.xml\nlogback:\n  loggers:\n    a: INFO\n"), 0o644); err != nil {
		t.Fatalf("write create file: %v", err)
	}

	if !waitForPath(t, events, cfgPath, 3*time.Second) {
		t.Fatalf("did not receive create/write event for %s", cfgPath)
	}

	if err := os.Remove(cfgPath); err != nil {
		t.Fatalf("remove file: %v", err)
	}

	if !waitForPath(t, events, cfgPath, 3*time.Second) {
		t.Fatalf("did not receive remove event for %s", cfgPath)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("watcher returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("watcher did not stop after cancel")
	}
}

// waitForPath waits for a batch that contains the expected path.
func waitForPath(t *testing.T, events <-chan []string, expected string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case batch := <-events:
			for _, path := range batch {
				if path == expected {
					return true
				}
			}
		case <-deadline:
			return false
		}
	}
}

// TestDirectory_InvalidDirectory verifies error when directory does not exist.
func TestDirectory_InvalidDirectory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan []string, 1)
	err := Directory(ctx, "/nonexistent/directory/path", 50*time.Millisecond, func(paths []string) {
		events <- paths
	})
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

// TestDirectory_RapidEvents verifies multiple rapid events are batched together.
func TestDirectory_RapidEvents(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan []string, 4)
	errCh := make(chan error, 1)

	go func() {
		errCh <- Directory(ctx, dir, 100*time.Millisecond, func(paths []string) {
			cp := make([]string, len(paths))
			copy(cp, paths)
			events <- cp
		})
	}()

	time.Sleep(150 * time.Millisecond)

	// Rapidly create multiple files
	files := []string{"a.yaml", "b.yaml", "c.yaml"}
	for _, name := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	// Should receive batched events
	select {
	case batch := <-events:
		if len(batch) < 1 {
			t.Fatal("expected batch of events")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive batched events")
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("watcher returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("watcher did not stop after cancel")
	}
}

// TestDirectory_ContextCancel verifies watcher stops on context cancellation.
func TestDirectory_ContextCancel(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())

	events := make(chan []string, 2)
	errCh := make(chan error, 1)

	go func() {
		errCh <- Directory(ctx, dir, 50*time.Millisecond, func(paths []string) {
			events <- paths
		})
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected no error on context cancel, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not stop after context cancel")
	}
}

// TestDirectory_SortedBatch verifies events are emitted in sorted order.
func TestDirectory_SortedBatch(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan []string, 2)
	errCh := make(chan error, 1)

	go func() {
		errCh <- Directory(ctx, dir, 100*time.Millisecond, func(paths []string) {
			cp := make([]string, len(paths))
			copy(cp, paths)
			events <- cp
		})
	}()

	time.Sleep(150 * time.Millisecond)

	// Create files in non-sorted order
	filenames := []string{"zebra.yaml", "apple.yaml", "middle.yaml"}
	for _, name := range filenames {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	// Receive batch
	select {
	case batch := <-events:
		// Verify batch is sorted
		for i := 1; i < len(batch); i++ {
			if batch[i] < batch[i-1] {
				t.Errorf("expected sorted batch, got: %v", batch)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive batch")
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("watcher returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("watcher did not stop")
	}
}
