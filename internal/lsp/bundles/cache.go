package bundles

import (
	"bytes"
	//nolint:gosec
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/util"

	rio "github.com/styrainc/regal/internal/io"
)

// Cache is a struct that maintains a number of bundles in memory and
// provides a way to refresh them when the source files change.
type Cache struct {
	errorLog      io.Writer
	bundles       map[string]*cacheBundle
	workspacePath string
}

type CacheOptions struct {
	ErrorLog      io.Writer
	WorkspacePath string
}

func NewCache(opts *CacheOptions) *Cache {
	workspacePath := opts.WorkspacePath

	if !strings.HasSuffix(workspacePath, rio.PathSeparator) {
		workspacePath += rio.PathSeparator
	}

	c := &Cache{
		workspacePath: workspacePath,
		bundles:       make(map[string]*cacheBundle),
	}

	if opts.ErrorLog != nil {
		c.errorLog = opts.ErrorLog
	}

	return c
}

// Refresh walks the workspace path and loads or refreshes any bundles that
// have changed since the last refresh.
func (c *Cache) Refresh() ([]string, error) {
	if c.workspacePath == "" {
		return nil, errors.New("workspace path is empty")
	}

	// find all the bundle roots that are currently present on disk
	foundBundleRoots, err := rio.FindManifestLocations(c.workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to walk workspace path: %w", err)
	}

	var refreshedBundles []string

	// refresh any bundles that have changed
	for _, root := range foundBundleRoots {
		if _, ok := c.bundles[root]; !ok {
			c.bundles[root] = &cacheBundle{}
		}

		refreshed, err := c.bundles[root].Refresh(filepath.Join(c.workspacePath, root))
		if err != nil {
			if c.errorLog != nil {
				fmt.Fprintf(c.errorLog, "failed to refresh bundle %q: %v\n", root, err)
			}

			continue
		}

		if refreshed {
			refreshedBundles = append(refreshedBundles, root)
		}
	}

	// remove any bundles that are no longer present on disk
	for root := range c.bundles {
		found := slices.Contains(foundBundleRoots, root)

		if !found {
			delete(c.bundles, root)
		}
	}

	return refreshedBundles, nil
}

// List returns a list of all the bundle roots that are currently present in
// the cache.
func (c *Cache) List() []string {
	return util.Keys(c.bundles)
}

// Get returns the bundle for the given root from the cache.
func (c *Cache) Get(root string) (bundle.Bundle, bool) {
	b, ok := c.bundles[root]
	if !ok {
		return bundle.Bundle{}, false
	}

	return b.bundle, true
}

// All returns all the bundles in the cache.
func (c *Cache) All() map[string]bundle.Bundle {
	bundles := make(map[string]bundle.Bundle, len(c.bundles))

	for root, cacheBundle := range c.bundles {
		bundles[root] = cacheBundle.bundle
	}

	return bundles
}

// cacheBundle is an internal struct that holds a bundle.Bundle and the MD5
// hash of each source file in the bundle. Hashes are used to determine if
// the bundle should be reloaded.
type cacheBundle struct {
	sourceDigests map[string][]byte
	bundle        bundle.Bundle
}

// Refresh loads the bundle from disk and updates the cache if any of the
// source files have changed since the last refresh.
func (c *cacheBundle) Refresh(path string) (bool, error) {
	// walk the bundle path and calculate the current MD5 hash of each file on disk
	walker := rio.NewFileWalkReducer(path, make(map[string][]byte)).
		WithSkipFunc(rio.DefaultSkipDirectories).
		WithFilters(
			rio.DirectoryFilter,
			rio.NegateFilter(rio.FileNameFilter(".manifest", "data.json", "data.yml", "data.yaml")),
		)

	onDiskDigests, err := walker.Reduce(func(p string, curr map[string][]byte) (map[string][]byte, error) {
		hash, err := calculateMD5(p)
		if err == nil {
			curr[p] = hash
		}

		return curr, err
	})
	if err != nil {
		return false, fmt.Errorf("failed to walk bundle path %q: %w", path, err)
	}

	// compare the files on disk with the files that have been seen before
	// and return without reloading the bundle if there have been no changes
	if maps.EqualFunc(c.sourceDigests, onDiskDigests, bytes.Equal) {
		return false, nil
	}

	// update the bundle's sourceDigests to the new on-disk state after a
	// successful refresh
	c.sourceDigests = onDiskDigests

	// if there has been any change in any of the source files, then
	// reload the bundle
	if c.bundle, err = LoadDataBundle(path); err != nil {
		return false, fmt.Errorf("failed to load bundle %q: %w", path, err)
	}

	return true, nil
}

func calculateMD5(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer file.Close()

	//nolint:gosec
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to calculate MD5 hash for file %q: %w", filePath, err)
	}

	return hash.Sum(nil), nil
}
