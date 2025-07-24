package examples

import (
	"sync"

	"github.com/styrainc/regal/pkg/roast/encoding"

	_ "embed"
)

type indexData struct {
	BuiltIns map[string]string `json:"builtins"`
	Keywords map[string]string `json:"keywords"`
}

const baseURL = "https://docs.styra.com/opa/rego-by-example"

var (
	//go:embed index.json
	indexJSON []byte
	indexOnce = sync.OnceValue(createIndex)
)

func createIndex() *indexData {
	index := &indexData{}

	if err := encoding.JSON().Unmarshal(indexJSON, index); err != nil {
		panic("failed to unmarshal built-in index: " + err.Error())
	}

	return index
}

// GetBuiltInLink returns the URL for the built-in function documentation
// if it has been documented, otherwise it returns false.
func GetBuiltInLink(builtinName string) (string, bool) {
	index := indexOnce()

	path, ok := index.BuiltIns[builtinName]
	if ok {
		return baseURL + "/" + path, true
	}

	return "", false
}

// GetKeywordLink returns the URL for the keyword documentation
// if it has been documented, otherwise it returns false.
func GetKeywordLink(keyword string) (string, bool) {
	index := indexOnce()

	path, ok := index.Keywords[keyword]
	if ok {
		return baseURL + "/" + path, true
	}

	return "", false
}
