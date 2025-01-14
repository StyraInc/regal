package fileprovider

import (
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
)

func TestCacheFileProvider(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	c.SetFileContents("file:///tmp/foo.rego", "package foo")
	c.SetFileContents("file:///tmp/bar.rego", "package bar")

	cfp := NewCacheFileProvider(c, clients.IdentifierGeneric)

	err := cfp.Put("file:///tmp/foo.rego", "package wow")
	if err != nil {
		t.Fatalf("failed to put file: %s", err)
	}

	contents, err := cfp.Get("file:///tmp/foo.rego")
	if err != nil {
		t.Fatalf("failed to get file: %s", err)
	}

	if contents != "package wow" {
		t.Fatalf("expected %s, got %s", "package wow", contents)
	}

	contentsStr, ok := c.GetFileContents("file:///tmp/foo.rego")
	if !ok {
		t.Fatalf("failed to get file contents")
	}

	if contentsStr != "package wow" {
		t.Fatalf("expected %s, got %s", "package wow", contents)
	}

	err = cfp.Rename("file:///tmp/foo.rego", "file:///tmp/wow.rego")
	if err != nil {
		t.Fatalf("failed to rename file: %s", err)
	}

	if _, ok := cfp.deletedFiles["file:///tmp/foo.rego"]; !ok {
		t.Fatalf("expected file to be deleted")
	}

	if _, ok := cfp.modifiedFiles["file:///tmp/wow.rego"]; !ok {
		t.Fatalf("expected file to be modified")
	}

	contents, err = cfp.Get("file:///tmp/wow.rego")
	if err != nil {
		t.Fatalf("failed to get file: %s", err)
	}

	if contents != "package wow" {
		t.Fatalf("expected %s, got %s", "package wow", contents)
	}
}
