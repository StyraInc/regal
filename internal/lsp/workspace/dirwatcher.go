package workspace

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// DirWatcherOpts holds options for creating a DirWatcher.
type DirWatcherOpts struct {
	ErrorLog        io.Writer
	PollingInterval time.Duration
	RootPath        string
}

// DirWatcher watches a directory for file changes, intended for monitoring
// a workspace directory for changes to files that are changed outwith
// the workspace's editor.
type DirWatcher struct {
	CreateChan chan string
	RemoveChan chan string
	UpdateChan chan string

	rootPath        string
	watcher         *fsnotify.Watcher
	errorLog        io.Writer
	pollingInterval time.Duration
	pathStore       map[string]struct{}
}

// NewDirWatcher creates a new DirWatcher with the given options.
func NewDirWatcher(opts *DirWatcherOpts) (*DirWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating watcher: %w", err)
	}

	dw := &DirWatcher{
		rootPath:        opts.RootPath,
		watcher:         watcher,
		errorLog:        opts.ErrorLog,
		pollingInterval: opts.PollingInterval,
		pathStore:       make(map[string]struct{}),

		CreateChan: make(chan string),
		RemoveChan: make(chan string),
		UpdateChan: make(chan string),
	}

	if dw.rootPath != "" {
		dw.addDir(dw.rootPath)
	}

	return dw, nil
}

func (dw *DirWatcher) Start(ctx context.Context) {
	go dw.watch(ctx)
	if dw.pollingInterval > 0 {
		go dw.poll(ctx)
	}
}

func (dw *DirWatcher) SetRootPath(path string) error {
	for _, path := range dw.watcher.WatchList() {
		dw.watcher.Remove(path)
	}

	err := dw.addDir(path)
	if err != nil {
		return fmt.Errorf("error adding directory %s: %w", path, err)
	}

	return nil
}

func (dw *DirWatcher) addDir(path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return dw.watcher.Add(path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error adding directory %s: %w", path, err)
	}

	return nil
}

func (dw *DirWatcher) watch(ctx context.Context) {
	ready := make(chan struct{})
	go func() {
		for dw.rootPath == "" {
			time.Sleep(300 * time.Millisecond)
		}
		ready <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return
	case <-ready:
	}

	fmt.Fprintln(dw.errorLog, "watcher starting")

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-dw.watcher.Events:
			fmt.Fprintln(dw.errorLog, "event", event.Name)
			if !ok {
				break
			}

			if event.Has(fsnotify.Remove) {
				_ = dw.watcher.Remove(event.Name)

				dw.RemoveChan <- event.Name
			}

			info, err := os.Stat(event.Name)
			if err != nil {
				break
			}

			if event.Has(fsnotify.Create) {
				if info.IsDir() {
					err := dw.addDir(event.Name)
					if err != nil {
						fmt.Fprintln(dw.errorLog, "Error:", err)
					}
				} else {
					dw.CreateChan <- event.Name
				}
			}

			if event.Has(fsnotify.Write) {
				if !info.IsDir() {
					dw.UpdateChan <- event.Name
				}
			}
		case err, ok := <-dw.watcher.Errors:
			if !ok {
				return
			}

			fmt.Fprintln(dw.errorLog, "Error:", err)
		}
	}
}

func (dw *DirWatcher) poll(ctx context.Context) {
	ready := make(chan struct{})
	go func() {
		for dw.rootPath == "" {
			time.Sleep(300 * time.Millisecond)
		}
		ready <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return
	case <-ready:
	}

	fmt.Fprintln(dw.errorLog, "poll starting")

	ticker := time.NewTicker(dw.pollingInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprintln(dw.errorLog, "poll run root", dw.rootPath)
			seenPaths := make(map[string]struct{})

			err := filepath.Walk(dw.rootPath, func(path string, info os.FileInfo, err error) error {
				if info == nil || info.IsDir() {
					return nil
				}

				seenPaths[path] = struct{}{}

				return nil
			})
			if err != nil {
				fmt.Fprintln(dw.errorLog, "Error:", err)
			}

			for path := range seenPaths {
				if _, ok := dw.pathStore[path]; !ok {
					dw.CreateChan <- path
					dw.pathStore[path] = struct{}{}
				}
			}

			for path := range dw.pathStore {
				if _, ok := dw.pathStore[path]; ok {
					dw.RemoveChan <- path
					delete(dw.pathStore, path)
				}
			}
		}
	}
}
