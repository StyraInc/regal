package fileprovider

import (
	"fmt"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/pkg/rules"
)

type FileProvider interface {
	List() ([]string, error)

	Get(string) (string, error)
	Put(string, string) error
	Delete(string) error
	Rename(string, string) error

	ToInput(versionsMap map[string]ast.RegoVersion) (rules.Input, error)
}

type RenameConflictError struct {
	From string
	To   string
}

func (e RenameConflictError) Error() string {
	return fmt.Sprintf("rename conflict: %q cannot be renamed as the target location %q already exists", e.From, e.To)
}
