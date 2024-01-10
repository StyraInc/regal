package lsp

import "os"

type StdOutReadWriteCloser struct{}

func (StdOutReadWriteCloser) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (c StdOutReadWriteCloser) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (c StdOutReadWriteCloser) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
