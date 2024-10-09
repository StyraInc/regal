package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/styrainc/regal/internal/lsp"
	"github.com/styrainc/regal/internal/lsp/log"
)

func init() {
	verboseLogging := false

	languageServerCommand := &cobra.Command{
		Use:   "language-server",
		Short: "Run the Regal Language Server",
		Long:  `Start the Regal Language Server and listen on stdin/stdout for client editor messages.`,

		RunE: wrapProfiling(func([]string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			opts := &lsp.LanguageServerOptions{
				LogWriter: os.Stderr,
				LogLevel:  log.LevelMessage,
			}

			ls := lsp.NewLanguageServer(ctx, opts)

			conn := lsp.NewConnectionFromLanguageServer(ctx, ls.Handle, &lsp.ConnectionOptions{
				LoggingConfig: lsp.ConnectionLoggingConfig{
					Writer:      os.Stderr,
					LogInbound:  verboseLogging,
					LogOutbound: verboseLogging,
				},
			})
			defer conn.Close()

			ls.SetConn(conn)
			go ls.StartDiagnosticsWorker(ctx)
			go ls.StartHoverWorker(ctx)
			go ls.StartCommandWorker(ctx)
			go ls.StartConfigWorker(ctx)
			go ls.StartWorkspaceStateWorker(ctx)
			go ls.StartTemplateWorker(ctx)
			go ls.StartWebServer(ctx)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			select {
			case <-conn.DisconnectNotify():
				fmt.Fprintln(os.Stderr, "Connection closed")
			case sig := <-sigChan:
				fmt.Fprintln(os.Stderr, "signal: ", sig.String())
			}

			return nil
		}),
	}

	languageServerCommand.Flags().BoolVarP(&verboseLogging, "verbose", "v", verboseLogging, "Enable verbose logging")

	addPprofFlag(languageServerCommand.Flags())

	RootCommand.AddCommand(languageServerCommand)
}
