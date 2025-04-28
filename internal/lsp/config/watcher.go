package config

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/styrainc/regal/internal/lsp/log"
)

type Watcher struct {
	logFunc func(log.Level, string, ...any)
	Reload  chan string
	Drop    chan struct{}

	pathUpdates chan string

	fsWatcher *fsnotify.Watcher

	path          string
	fsWatcherLock sync.Mutex
}

type WatcherOpts struct {
	LogFunc func(log.Level, string, ...any)
	Path    string
}

func NewWatcher(opts *WatcherOpts) *Watcher {
	w := &Watcher{
		Reload:      make(chan string, 1),
		Drop:        make(chan struct{}, 1),
		pathUpdates: make(chan string, 1),
	}

	if opts != nil {
		w.logFunc = opts.LogFunc
		w.path = opts.Path
	}

	return w
}

func (w *Watcher) Start(ctx context.Context) error {
	err := w.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop existing watcher: %w", err)
	}

	w.fsWatcherLock.Lock()
	w.fsWatcher, err = fsnotify.NewWatcher()
	w.fsWatcherLock.Unlock()

	if err != nil {
		return fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	go func() {
		w.loop(ctx)
	}()

	return nil
}

func (w *Watcher) Watch(configFilePath string) {
	w.pathUpdates <- configFilePath
}

func (w *Watcher) Stop() error {
	if w.fsWatcher != nil {
		if err := w.fsWatcher.Close(); err != nil {
			return fmt.Errorf("failed to close fsnotify watcher: %w", err)
		}

		return nil
	}

	return nil
}

func (w *Watcher) IsWatching() bool {
	w.fsWatcherLock.Lock()
	defer w.fsWatcherLock.Unlock()

	if w.fsWatcher == nil {
		return false
	}

	return len(w.fsWatcher.WatchList()) > 0
}

func (w *Watcher) loop(ctx context.Context) {
	for {
		select {
		case path := <-w.pathUpdates:
			if w.path != "" {
				err := w.fsWatcher.Remove(w.path)
				if err != nil {
					w.logFunc(log.LevelMessage, "failed to remove existing watch: %v\n", err)
				}
			}

			if err := w.fsWatcher.Add(path); err != nil {
				w.logFunc(log.LevelDebug, "failed to add watch: %v\n", err)
			}

			w.path = path

			// when the path itself is changed, then this is an event too
			w.Reload <- path
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				w.logFunc(log.LevelMessage, "config watcher event channel closed\n")

				return
			}

			if event.Has(fsnotify.Write) {
				w.Reload <- event.Name
			}

			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				w.path = ""
				w.Drop <- struct{}{}
			}
		case err := <-w.fsWatcher.Errors:
			w.logFunc(log.LevelMessage, "config watcher error: %v\n", err)
		case <-ctx.Done():
			if err := w.Stop(); err != nil {
				w.logFunc(log.LevelMessage, "failed to stop watcher: %v\n", err)
			}

			return
		}
	}
}
