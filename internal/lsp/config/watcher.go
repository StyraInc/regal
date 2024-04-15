package config

import (
	"context"
	"fmt"
	"io"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	Reload chan string
	Drop   chan struct{}

	path        string
	pathUpdates chan string

	fsWatcher *fsnotify.Watcher

	errorWriter io.Writer
}

type WatcherOpts struct {
	ErrorWriter io.Writer
	Path        string
}

func NewWatcher(opts *WatcherOpts) *Watcher {
	w := &Watcher{
		Reload:      make(chan string, 1),
		Drop:        make(chan struct{}, 1),
		pathUpdates: make(chan string, 1),
	}

	if opts != nil {
		w.errorWriter = opts.ErrorWriter
		w.path = opts.Path
	}

	return w
}

func (w *Watcher) Start(ctx context.Context) error {
	err := w.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop existing watcher: %w", err)
	}

	w.fsWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	go func() {
		w.loop(ctx)
	}()

	return nil
}

func (w *Watcher) loop(ctx context.Context) {
	for {
		select {
		case path := <-w.pathUpdates:
			if w.path != "" {
				err := w.fsWatcher.Remove(w.path)
				if err != nil {
					fmt.Fprintf(w.errorWriter, "failed to remove existing watch: %v\n", err)
				}
			}

			err := w.fsWatcher.Add(path)
			if err != nil {
				fmt.Fprintf(w.errorWriter, "failed to add watch: %v\n", err)
			}

			w.path = path

			// when the path itself is changed, then this is an event too
			w.Reload <- path
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				fmt.Fprintf(w.errorWriter, "config watcher event channel closed\n")

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
			fmt.Fprintf(w.errorWriter, "config watcher error: %v\n", err)
		case <-ctx.Done():
			err := w.Stop()
			if err != nil {
				fmt.Fprintf(w.errorWriter, "failed to stop watcher: %v\n", err)
			}

			return
		}
	}
}

func (w *Watcher) Watch(configFilePath string) {
	w.pathUpdates <- configFilePath
}

func (w *Watcher) Stop() error {
	if w.fsWatcher != nil {
		err := w.fsWatcher.Close()
		if err != nil {
			return fmt.Errorf("failed to close fsnotify watcher: %w", err)
		}

		return nil
	}

	return nil
}

func (w *Watcher) IsWatching() bool {
	if w.fsWatcher == nil {
		return false
	}

	return len(w.fsWatcher.WatchList()) > 0
}
