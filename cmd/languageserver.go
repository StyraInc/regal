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

		RunE: wrapProfiling(func(args []string) error {

			ctx, cancel := context.WithCancel(context.Background())

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
			ls.StartDiagnosticsWorker(ctx)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			select {
			case <-conn.DisconnectNotify():
				fmt.Fprint(os.Stderr, "Connection closed\n")
				cancel()
			case sig := <-sigChan:
				fmt.Fprint(os.Stderr, "signal: ", sig.String(), "\n")
				cancel()
			}

			return nil
		}),
	}

	RootCommand.AddCommand(languageServerCommand)
}
