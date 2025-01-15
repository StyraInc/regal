package fileprovider

import (
	"fmt"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/rules"
)

type CacheFileProvider struct {
	Cache            *cache.Cache
	ClientIdentifier clients.Identifier

	modifiedFiles map[string]struct{}
	deletedFiles  map[string]struct{}
}

func NewCacheFileProvider(c *cache.Cache, ci clients.Identifier) *CacheFileProvider {
	return &CacheFileProvider{
		Cache:            c,
		ClientIdentifier: ci,
		modifiedFiles:    make(map[string]struct{}),
		deletedFiles:     make(map[string]struct{}),
	}
}

func (c *CacheFileProvider) List() ([]string, error) {
	uris := util.Keys(c.Cache.GetAllFiles())

	paths := make([]string, len(uris))
	for i, u := range uris {
		paths[i] = uri.ToPath(c.ClientIdentifier, u)
	}

	return paths, nil
}

func (c *CacheFileProvider) Get(file string) (string, error) {
	contents, ok := c.Cache.GetFileContents(uri.FromPath(c.ClientIdentifier, file))
	if !ok {
		return "", fmt.Errorf("failed to get file %s", file)
	}

	return contents, nil
}

func (c *CacheFileProvider) Put(file string, content string) error {
	c.Cache.SetFileContents(file, content)

	return nil
}

func (c *CacheFileProvider) Delete(file string) error {
	c.Cache.Delete(uri.FromPath(c.ClientIdentifier, file))

	return nil
}

func (c *CacheFileProvider) Rename(from, to string) error {
	fromURI := uri.FromPath(c.ClientIdentifier, from)
	toURI := uri.FromPath(c.ClientIdentifier, to)

	content, ok := c.Cache.GetFileContents(fromURI)
	if !ok {
		return fmt.Errorf("file %s not found", from)
	}

	if _, exists := c.Cache.GetFileContents(toURI); exists {
		return RenameConflictError{
			From: from,
			To:   to,
		}
	}

	c.Cache.SetFileContents(toURI, content)

	c.modifiedFiles[to] = struct{}{}

	c.Cache.Delete(fromURI)

	delete(c.modifiedFiles, from)
	c.deletedFiles[from] = struct{}{}

	return nil
}

func (c *CacheFileProvider) ToInput(versionsMap map[string]ast.RegoVersion) (rules.Input, error) {
	input, err := rules.InputFromMap(c.Cache.GetAllFiles(), versionsMap)
	if err != nil {
		return rules.Input{}, fmt.Errorf("failed to create input: %w", err)
	}

	return input, nil
}
