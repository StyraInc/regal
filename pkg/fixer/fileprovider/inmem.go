package fileprovider

import (
	"fmt"
	"os"

	rutil "github.com/anderseknert/roast/pkg/util"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/rules"
)

type InMemoryFileProvider struct {
	files         map[string][]byte
	modifiedFiles map[string]struct{}
	deletedFiles  map[string]struct{}
}

func NewInMemoryFileProvider(files map[string][]byte) *InMemoryFileProvider {
	return &InMemoryFileProvider{
		files:         files,
		modifiedFiles: make(map[string]struct{}),
		deletedFiles:  make(map[string]struct{}),
	}
}

func NewInMemoryFileProviderFromFS(paths ...string) (*InMemoryFileProvider, error) {
	files := make(map[string][]byte)

	for _, path := range paths {
		fc, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", path, err)
		}

		files[path] = fc
	}

	return &InMemoryFileProvider{
		files:         files,
		modifiedFiles: make(map[string]struct{}),
		deletedFiles:  make(map[string]struct{}),
	}, nil
}

func (p *InMemoryFileProvider) List() ([]string, error) {
	files := make([]string, 0)
	for file := range p.files {
		files = append(files, file)
	}

	return files, nil
}

func (p *InMemoryFileProvider) Get(file string) ([]byte, error) {
	content, ok := p.files[file]
	if !ok {
		return nil, fmt.Errorf("file %s not found", file)
	}

	return content, nil
}

func (p *InMemoryFileProvider) Put(file string, content []byte) error {
	p.files[file] = content

	p.modifiedFiles[file] = struct{}{}

	return nil
}

func (p *InMemoryFileProvider) Rename(from, to string) error {
	content, ok := p.files[from]
	if !ok {
		return fmt.Errorf("file %s not found", from)
	}

	_, ok = p.files[to]
	if ok {
		return RenameConflictError{
			From: from,
			To:   to,
		}
	}

	if err := p.Put(to, content); err != nil {
		return fmt.Errorf("failed to put file %s: %w", to, err)
	}

	if err := p.Delete(from); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", from, err)
	}

	return nil
}

func (p *InMemoryFileProvider) Delete(file string) error {
	p.deletedFiles[file] = struct{}{}

	delete(p.files, file)
	delete(p.modifiedFiles, file)

	return nil
}

func (p *InMemoryFileProvider) ModifiedFiles() []string {
	return util.Keys(p.modifiedFiles)
}

func (p *InMemoryFileProvider) DeletedFiles() []string {
	return util.Keys(p.deletedFiles)
}

// TODO: We need a way to specify the Rego version for the files here and avoid
// relying on the parser to infer those.
func (p *InMemoryFileProvider) ToInput() (rules.Input, error) {
	strContents := make(map[string]string)
	modules := make(map[string]*ast.Module)

	for filename, content := range p.files {
		var err error

		strContents[filename] = rutil.ByteSliceToString(content)

		modules[filename], err = parse.ModuleWithOpts(
			filename,
			strContents[filename],
			parse.ParserOptions(),
		)
		if err != nil {
			return rules.Input{}, fmt.Errorf("failed to parse module %s: %w", filename, err)
		}
	}

	return rules.NewInput(strContents, modules), nil
}
