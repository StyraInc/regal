import { window, workspace, ExtensionContext } from 'vscode';

import {
    ErrorHandlerResult,
    CloseHandlerResult,
    ErrorAction,
    Message,
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    CloseAction
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: ExtensionContext) {
    const serverOptions: ServerOptions = {
        command: "regal",
        args: ["language-server"],
    };

    const outChan = window.createOutputChannel("regal-ls");

    const clientOptions: LanguageClientOptions = {
      documentSelector: [{ scheme: 'file', language: 'rego' }],
      outputChannel: outChan,
      traceOutputChannel: outChan,
      revealOutputChannelOn: 0,
      errorHandler: {
        error: (error: Error, message: Message, count: number): ErrorHandlerResult => {
          console.error(error);
          console.error(message);
          return {
            action: ErrorAction.Continue,
          };
        },
        closed: (): CloseHandlerResult => {
          console.error("client closed");
          return {
            action: CloseAction.DoNotRestart,
          };
        },
      },
      synchronize: {
        fileEvents: [
          workspace.createFileSystemWatcher('**/*.rego'),
          workspace.createFileSystemWatcher('**/.regal/config.yaml'),
        ],
      },
      diagnosticPullOptions: {
        onChange: true,
        onSave: true,
      },
    };

    client = new LanguageClient(
        'regal',
        'Regal LSP client',
        serverOptions,
        clientOptions
    );

    client.start();
}

export function deactivate(): Thenable<void> | undefined {
  console.log("deactivating");
  if (!client) {
    return undefined;
  }
  return client.stop();
}
