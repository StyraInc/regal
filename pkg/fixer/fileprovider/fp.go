package fileprovider

import "github.com/styrainc/regal/pkg/rules"

type FileProvider interface {
	List() ([]string, error)

	Get(string) ([]byte, error)
	Put(string, []byte) error
	Delete(string) error
	Rename(string, string) error

	ToInput() (rules.Input, error)
}
