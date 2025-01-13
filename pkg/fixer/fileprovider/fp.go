package fileprovider

import (
	"fmt"

	"github.com/styrainc/regal/pkg/rules"
)

type FileProvider interface {
	List() ([]string, error)

	Get(string) ([]byte, error)
	Put(string, []byte) error
	Delete(string) error
	Rename(string, string) error

	ToInput() (rules.Input, error)
}

type RenameConflictError struct {
	From string
	To   string
}

func (e RenameConflictError) Error() string {
	return fmt.Sprintf("rename conflict: %q cannot be renamed as the target location %q already exists", e.From, e.To)
}
