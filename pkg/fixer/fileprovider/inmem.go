package fileprovider

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/rules"
)

type InMemoryFileProvider struct {
	files map[string][]byte
}

func NewInMemoryFileProvider(files map[string][]byte) *InMemoryFileProvider {
	return &InMemoryFileProvider{
		files: files,
	}
}

func (p *InMemoryFileProvider) ListFiles() ([]string, error) {
	files := make([]string, 0)
	for file := range p.files {
		files = append(files, file)
	}

	return files, nil
}

func (p *InMemoryFileProvider) GetFile(file string) ([]byte, error) {
	content, ok := p.files[file]
	if !ok {
		return nil, fmt.Errorf("file %s not found", file)
	}

	return content, nil
}

func (p *InMemoryFileProvider) PutFile(file string, content []byte) error {
	p.files[file] = content

	return nil
}

func (p *InMemoryFileProvider) ToInput() (rules.Input, error) {
	modules := make(map[string]*ast.Module)

	for filename, content := range p.files {
		var err error

		modules[filename], err = parse.Module(filename, string(content))
		if err != nil {
			return rules.Input{}, fmt.Errorf("failed to parse module %s: %w", filename, err)
		}
	}

	strContents := make(map[string]string)
	for filename, content := range p.files {
		strContents[filename] = string(content)
	}

	return rules.NewInput(strContents, modules), nil
}
