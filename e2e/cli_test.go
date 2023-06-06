//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCLIUsage(t *testing.T) {
	t.Parallel()

	if err := regal()(); err != nil {
		t.Fatal(err)
	}
}

func TestLintEmptyDir(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		format string
		check  func(*testing.T, *bytes.Buffer)
	}{
		{
			format: "pretty",
			check: func(t *testing.T, out *bytes.Buffer) {
				t.Helper()
				if exp, act := "0 files linted. No violations found.\n", out.String(); exp != act {
					t.Errorf("output: expected %q, got %q", exp, act)
				}
			},
		},
		{
			format: "compact",
			check: func(t *testing.T, out *bytes.Buffer) {
				t.Helper()
				if exp, act := "\n", out.String(); exp != act {
					t.Errorf("output: expected %q, got %q", exp, act)
				}
			},
		},
		{
			format: "json",
			check: func(t *testing.T, out *bytes.Buffer) {
				t.Helper()
				s := struct {
					Violations []string       `json:"violations"`
					Summary    map[string]any `json:"summary"`
				}{}
				if err := json.NewDecoder(out).Decode(&s); err != nil {
					t.Fatal(err)
				}
				if exp, act := 0, len(s.Violations); exp != act {
					t.Errorf("violations: expected %d, got %d", exp, act)
				}
				zero := float64(0)
				exp := map[string]any{"files_scanned": zero, "files_failed": zero, "files_skipped": zero, "num_violations": zero}
				if diff := cmp.Diff(exp, s.Summary); diff != "" {
					t.Errorf("unexpected summary (-want, +got):\n%s", diff)
				}
			},
		},
	} {
		tc := tc
		t.Run(tc.format, func(t *testing.T) {
			t.Parallel()

			out := bytes.Buffer{}

			err := regal(&out)("lint", "--format", tc.format, t.TempDir())
			if err != nil {
				t.Fatalf("%v %[1]T", err)
			}

			tc.check(t, &out)
		})
	}
}

func TestLintNonExistantDir(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	td := t.TempDir()

	err := regal(&stdout, &stderr)("lint", td+"/what/ever")
	if exp, act := 1, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	if exp, act := "errors encountered when reading files to lint: failed to load policy from provided args: "+
		"1 error occurred during loading: stat "+td+"/what/ever: no such file or directory\n", stdout.String(); exp != act {
		t.Errorf("expected stdout %q, got %q", exp, act)
	}
}

func binary() string {
	if b := os.Getenv("REGAL_BIN"); b != "" {
		return b
	}

	return "../regal"
}

func regal(outs ...io.Writer) func(...string) error {
	return func(args ...string) error {
		c := exec.Command(binary(), args...)

		if len(outs) > 0 {
			c.Stdout = outs[0]
		}

		if len(outs) > 1 {
			c.Stderr = outs[0]
		}

		return c.Run() //nolint:wrapcheck // We're in tests. This is fine.
	}
}

type exitStatus interface {
	ExitStatus() int
}

// ExitStatus returns the exit status of the error if it is an exec.ExitError
// or if it implements ExitStatus() int.
// 0 if it is nil or panics if it is a different error.
func ExitStatus(err error) int {
	switch e := err.(type) { //nolint:errorlint // We know the errors that can happen here, the switch is enough.
	case nil:
		return 0
	case exitStatus:
		return e.ExitStatus()
	case *exec.ExitError:
		if ex, ok := e.Sys().(exitStatus); ok {
			return ex.ExitStatus()
		}
	}

	panic("unreachable")
}
