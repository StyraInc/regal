package examples

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// GetBuiltInLink returns the URL for the built-in function documentation
// if it has been documented, otherwise it returns false.
func GetBuiltInLink(builtinName string) (string, bool) {
	path, ok := index.BuiltIns[builtinName]
	if ok {
		return fmt.Sprintf("%s/%s", baseURL, path), true
	}

	return "", false
}

// GetKeywordLink returns the URL for the keyword documentation
// if it has been documented, otherwise it returns false.
func GetKeywordLink(keyword string) (string, bool) {
	path, ok := index.Keywords[keyword]
	if ok {
		return fmt.Sprintf("%s/%s", baseURL, path), true
	}

	return "", false
}

//go:embed index.json
var indexJSON []byte

type indexData struct {
	BuiltIns map[string]string `json:"builtins"`
	Keywords map[string]string `json:"keywords"`
}

const baseURL = "https://docs.styra.com/opa/rego-by-example"

var index *indexData

func init() {
	index = &indexData{}

	if err := json.Unmarshal(indexJSON, index); err != nil {
		panic("failed to unmarshal built-in index: " + err.Error())
	}
}
