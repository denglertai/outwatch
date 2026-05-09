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
