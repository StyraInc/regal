package lsp

import (
	"fmt"
	"os"
)

type StdOutReadWriteCloser struct{}

func (StdOutReadWriteCloser) Read(p []byte) (int, error) {
	c, err := os.Stdin.Read(p)

	if err != nil {
		return c, fmt.Errorf("failed to read from stdin: %w", err)
	}

	return c, nil
}

func (StdOutReadWriteCloser) Write(p []byte) (int, error) {
	c, err := os.Stdout.Write(p)

	if err != nil {
		return c, fmt.Errorf("failed to write to stdout: %w", err)
	}

	return c, nil
}

func (StdOutReadWriteCloser) Close() error {
	return nil
}
