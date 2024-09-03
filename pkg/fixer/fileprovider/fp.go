package fileprovider

import "github.com/styrainc/regal/pkg/rules"

type FileProvider interface {
	ListFiles() ([]string, error)
	GetFile(string) ([]byte, error)
	PutFile(string, []byte) error
	DeleteFile(string) error
	ToInput() (rules.Input, error)
}
