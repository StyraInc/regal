package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/styrainc/regal/internal/lsp"
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
				ErrorLog: os.Stderr,
			}

			ls := lsp.NewLanguageServer(opts)

			loggingCfg := lsp.ConnectionLoggingConfig{
				Writer:      os.Stderr,
				LogInbound:  false,
				LogOutbound: false,
			}
			if verboseLogging {
				loggingCfg.LogInbound = true
				loggingCfg.LogOutbound = true
			}

			conn := lsp.NewConnectionFromLanguageServer(ctx, ls.Handle, &lsp.ConnectionOptions{
				LoggingConfig: loggingCfg,
			})

			ls.SetConn(conn)

			go ls.StartDiagnosticsWorker(ctx)
			go ls.StartHoverWorker(ctx)
			go ls.StartCommandWorker(ctx)

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

	languageServerCommand.Flags().BoolVarP(&verboseLogging, "verbose", "v", verboseLogging, "Enable verbose logging")

	RootCommand.AddCommand(languageServerCommand)
}
