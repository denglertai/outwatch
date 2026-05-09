// Package watch provides filesystem watch orchestration helpers.
package watch

import (
	"context"
	"sort"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Directory watches a directory and emits debounced change batches.
func Directory(ctx context.Context, dir string, debounce time.Duration, onBatch func([]string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(dir); err != nil {
		return err
	}

	pending := map[string]struct{}{}
	flush := func() {
		if len(pending) == 0 {
			return
		}
		paths := make([]string, 0, len(pending))
		for path := range pending {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		pending = map[string]struct{}{}
		onBatch(paths)
	}

	timer := time.NewTimer(debounce)
	if !timer.Stop() {
		<-timer.C
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-watcher.Errors:
			if err != nil {
				return err
			}
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			pending[event.Name] = struct{}{}
			timer.Reset(debounce)
		case <-timer.C:
			flush()
		}
	}
}
