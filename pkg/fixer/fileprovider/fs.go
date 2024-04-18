package fileprovider

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/rules"
)

type FSFileProvider struct {
	roots []string
}

func NewFSFileProvider(roots ...string) *FSFileProvider {
	return &FSFileProvider{
		roots: roots,
	}
}

func (p *FSFileProvider) ListFiles() ([]string, error) {
	var files []string

	for _, root := range p.roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				files = append(files, path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}
	}

	return files, nil
}

func (*FSFileProvider) GetFile(file string) ([]byte, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", file, err)
	}

	return content, nil
}

func (*FSFileProvider) PutFile(file string, content []byte) error {
	stat, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", file, err)
	}

	err = os.WriteFile(file, content, stat.Mode())
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", file, err)
	}

	return nil
}

func (p *FSFileProvider) ToInput() (rules.Input, error) {
	modules := make(map[string]*ast.Module)

	files, err := p.ListFiles()
	if err != nil {
		return rules.Input{}, fmt.Errorf("failed to list files: %w", err)
	}

	strContents := make(map[string]string)

	for _, filename := range files {
		content, err := p.GetFile(filename)
		if err != nil {
			return rules.Input{}, fmt.Errorf("failed to get file %s: %w", filename, err)
		}

		modules[filename], err = parse.Module(filename, string(content))
		if err != nil {
			return rules.Input{}, fmt.Errorf("failed to parse module %s: %w", filename, err)
		}

		strContents[filename] = string(content)
	}

	return rules.NewInput(strContents, modules), nil
}
