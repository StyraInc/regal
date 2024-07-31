package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestThings(t *testing.T) {
	for i := 1; i < 3; i++ {
		fmt.Println(i)
		mytest(t)
	}
}

func mytest(t *testing.T) {
	defaultTimeout := 80 * time.Millisecond

	rootDir := t.TempDir()

	fooDir := filepath.Join(rootDir, "foo")

	err := os.Mkdir(fooDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	barFile := filepath.Join(fooDir, "bar.txt")

	err = os.WriteFile(barFile, []byte("initial content"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	bazFile := filepath.Join(rootDir, "baz.txt")

	err = os.WriteFile(bazFile, []byte("initial content"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	opts := &DirWatcherOpts{
		ErrorLog:        os.Stdout,
		PollingInterval: 10 * time.Millisecond,
		RootPath:        rootDir,
	}

	dw, err := NewDirWatcher(opts)
	if err != nil {
		t.Fatal(err)
	}
	dw.Start(context.Background())

	// Test updating baz.txt
	err = os.WriteFile(bazFile, []byte("new content"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	waitForEvent(t, dw, "update", bazFile, defaultTimeout)

	// Test removing foo/bar.txt
	err = os.Remove(barFile)
	if err != nil {
		t.Fatal(err)
	}

	waitForEvent(t, dw, "remove", barFile, defaultTimeout)

	// Test creating foo/bax.txt
	baxFile := filepath.Join(fooDir, "bax.txt")

	err = os.WriteFile(baxFile, []byte("content"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	waitForEvent(t, dw, "create", baxFile, defaultTimeout)

	// Test creating new directory foo/bar/baz and file foo/bar/baz/quz.txt
	bazDir := filepath.Join(fooDir, "bar", "baz")

	err = os.MkdirAll(bazDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	quzFile := filepath.Join(bazDir, "quz.txt")

	err = os.WriteFile(quzFile, []byte("content"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	waitForEvent(t, dw, "create", quzFile, defaultTimeout)

	// Test removing a directory only sends events for files
	err = os.RemoveAll(bazDir)
	if err != nil {
		t.Fatal(err)
	}

	waitForEvent(t, dw, "remove", quzFile, defaultTimeout)
}

func waitForEvent(t *testing.T, dw *DirWatcher, expectedEvent string, expectedPath string, timeout time.Duration) {
	t.Helper()
	timer := time.After(timeout)

	for {
		select {
		case update := <-dw.UpdateChan:
			if expectedEvent == "update" && update == expectedPath {
				return
			}
		case remove := <-dw.RemoveChan:
			if expectedEvent == "remove" && remove == expectedPath {
				return
			}
		case create := <-dw.CreateChan:
			if expectedEvent == "create" && create == expectedPath {
				return
			}
		case <-timer:
			t.Fatalf("expected %s event for %s, but timed out", expectedEvent, expectedPath)
			return
		}
	}
}
