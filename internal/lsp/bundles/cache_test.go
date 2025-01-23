package bundles

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/styrainc/regal/internal/testutil"
)

func TestRefresh(t *testing.T) {
	t.Parallel()

	workspacePath := testutil.TempDirectoryOf(t, map[string]string{
		"foo/.manifest": `{"roots":["foo"]}`,
		"foo/data.json": `{"foo": "bar"}`,
	})

	c := NewCache(&CacheOptions{WorkspacePath: workspacePath})

	// perform the first load of the bundles
	refreshedBundles, err := c.Refresh()
	if err != nil {
		t.Fatalf("failed to refresh cache: %v", err)
	}

	if !slices.Equal(refreshedBundles, []string{"foo"}) {
		t.Fatalf("unexpected refreshed bundles: %v", refreshedBundles)
	}

	if len(c.List()) != 1 {
		t.Fatalf("unexpected number of bundles: %d", len(c.List()))
	}

	fooBundle, ok := c.Get("foo")
	if !ok {
		t.Fatalf("failed to get bundle foo")
	}

	if !reflect.DeepEqual(fooBundle.Data, map[string]any{"foo": "bar"}) {
		t.Fatalf("unexpected bundle data: %v", fooBundle.Data)
	}

	if fooBundle.Manifest.Roots == nil {
		t.Fatalf("unexpected bundle roots: %v", fooBundle.Manifest.Roots)
	}

	if !reflect.DeepEqual(*fooBundle.Manifest.Roots, []string{"foo"}) {
		t.Fatalf("unexpected bundle roots: %v", *fooBundle.Manifest.Roots)
	}

	// perform the second load of the bundles, after no changes on disk
	refreshedBundles, err = c.Refresh()
	if err != nil {
		t.Fatalf("failed to refresh cache: %v", err)
	}

	if !slices.Equal(refreshedBundles, []string{}) {
		t.Fatalf("unexpected refreshed bundles: %v", refreshedBundles)
	}

	// add a new unrelated file
	testutil.MustWriteFile(t, filepath.Join(workspacePath, "foo", "foo.rego"), []byte(`package wow`))

	// perform the third load of the bundles, after adding a new unrelated file
	refreshedBundles, err = c.Refresh()
	if err != nil {
		t.Fatalf("failed to refresh cache: %v", err)
	}

	if !slices.Equal(refreshedBundles, []string{}) {
		t.Fatalf("unexpected refreshed bundles: %v", refreshedBundles)
	}

	// update the data in the bundle
	testutil.MustWriteFile(t, filepath.Join(workspacePath, "foo", "data.json"), []byte(`{"foo": "baz"}`))

	refreshedBundles, err = c.Refresh()
	if err != nil {
		t.Fatalf("failed to refresh cache: %v", err)
	}

	if !slices.Equal(refreshedBundles, []string{"foo"}) {
		t.Fatalf("unexpected refreshed bundles: %v", refreshedBundles)
	}

	fooBundle, ok = c.Get("foo")
	if !ok {
		t.Fatalf("failed to get bundle foo")
	}

	if !reflect.DeepEqual(fooBundle.Data, map[string]any{"foo": "baz"}) {
		t.Fatalf("unexpected bundle data: %v", fooBundle.Data)
	}

	// create a new bundle
	testutil.MustMkdirAll(t, workspacePath, "bar")
	testutil.MustWriteFile(t, filepath.Join(workspacePath, "bar", ".manifest"), []byte(`{"roots":["bar"]}`))
	testutil.MustWriteFile(t, filepath.Join(workspacePath, "bar", "data.json"), []byte(`{"bar": true}`))

	refreshedBundles, err = c.Refresh()
	if err != nil {
		t.Fatalf("failed to refresh cache: %v", err)
	}

	if !slices.Equal(refreshedBundles, []string{"bar"}) {
		t.Fatalf("unexpected refreshed bundles: %v", refreshedBundles)
	}

	barBundle, ok := c.Get("bar")
	if !ok {
		t.Fatalf("failed to get bundle foo")
	}

	if !reflect.DeepEqual(barBundle.Data, map[string]any{"bar": true}) {
		t.Fatalf("unexpected bundle data: %v", fooBundle.Data)
	}

	// remove the foo bundle
	if err = os.RemoveAll(filepath.Join(workspacePath, "foo")); err != nil {
		t.Fatalf("failed to remove foo bundle: %v", err)
	}

	if _, err = c.Refresh(); err != nil {
		t.Fatalf("failed to refresh cache: %v", err)
	}

	if !slices.Equal(c.List(), []string{"bar"}) {
		t.Fatalf("unexpected bundle list: %v", c.List())
	}
}
