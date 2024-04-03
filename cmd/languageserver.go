package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/styrainc/regal/internal/lsp"
)

func init() {
	languageServerCommand := &cobra.Command{
		Use:   "language-server",
		Short: "Run the Regal Language Server",
		Long:  `Start the Regal Language Server and listen on stdin/stdout for client editor messages.`,

		RunE: wrapProfiling(func([]string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			opts := &lsp.LanguageServerOptions{
				ErrorLog:       os.Stderr,
				VerboseLogging: true,
			}

			ls := lsp.NewLanguageServer(opts)

			conn := jsonrpc2.NewConn(
				ctx,
				jsonrpc2.NewBufferedStream(lsp.StdOutReadWriteCloser{}, jsonrpc2.VSCodeObjectCodec{}),
				jsonrpc2.HandlerWithError(ls.Handle),
			)

			ls.SetConn(conn)
			go ls.StartDiagnosticsWorker(ctx)
			go ls.StartHoverWorker(ctx)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			select {
			case <-conn.DisconnectNotify():
				fmt.Fprint(os.Stderr, "Connection closed\n")
			case sig := <-sigChan:
				fmt.Fprint(os.Stderr, "signal: ", sig.String(), "\n")
			}

			return nil
		}),
	}

	RootCommand.AddCommand(languageServerCommand)
}
