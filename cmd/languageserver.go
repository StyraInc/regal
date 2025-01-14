package cmd

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/styrainc/regal/internal/lsp"
	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/pkg/version"
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

			exe, err := os.Executable()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error getting executable:", err)
			} else {
				absPath, err := filepath.Abs(exe)
				if err != nil {
					fmt.Fprintln(os.Stderr, "error getting executable path:", err)
				} else {
					fmt.Fprintf(
						os.Stderr,
						"Regal Language Server (path: %s, version: %s)",
						absPath,
						cmp.Or(version.Version, "Unknown"),
					)
				}
			}

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
