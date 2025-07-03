package main

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/lsp"
)

//go:embed *.html *.css
var root embed.FS

func main() {
	http.Handle("/regal", http.HandlerFunc(handleWS))
	http.Handle("/", http.FileServerFS(root))
	panic(http.ListenAndServe(":8787", nil))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true }, // TODO: check something
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	cfg := config.Config{
		Rules: map[string]config.Category{
			"idiomatic": {
				"directory-package-mismatch": config.Rule{
					Level: "ignore",
				},
			},
			"testing": {
				"test-outside-test-package": config.Rule{
					Level: "ignore",
				},
				"file-missing-test-suffix": config.Rule{
					Level: "ignore",
				},
			},
			"imports": {
				"unresolved-import": config.Rule{
					Level: "ignore",
				},
				"unresolved-reference": config.Rule{
					Level: "ignore",
				},
				"prefer-package-imports": config.Rule{
					Level: "error",
					Extra: map[string]any{
						"ignore-import-paths": []string{"data.system.eopa.utils.tests.v1.filter"},
					},
				},
			},
		},
	}

	ls, err := lsp.New(r.Context(), ws, &cfg) // ls takes ownership of ws
	if err != nil {
		slog.Error(err.Error())
		return
	}

	defer ls.Close() //nolint:errcheck
	if err := ls.Wait(r.Context()); err != nil {
		slog.Error(err.Error())
		return
	}
}
