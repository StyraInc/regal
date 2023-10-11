package cmd

import "fmt"

type ExitError struct {
	code int
}

func (e ExitError) Error() string {
	return fmt.Sprintf("exit code %d", e.code)
}

func (e ExitError) Code() int {
	return e.code
}

func exit(code int) error {
	return ExitError{code: code}
}
