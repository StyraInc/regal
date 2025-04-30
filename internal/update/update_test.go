package update

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCheckAndWarn(t *testing.T) {
	t.Parallel()

	remoteCalls := 0

	localReleasesServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if _, err := w.Write([]byte(`{"tag_name": "v0.2.0"}`)); err != nil {
				t.Fatal(err)
			}

			remoteCalls++
		}),
	)

	w := bytes.NewBuffer(nil)

	tempStateDir := t.TempDir()

	opts := Options{
		CurrentVersion: "v0.1.0",
		CurrentTime:    time.Now().UTC(),
		StateDir:       tempStateDir,

		ReleaseServerHost: localReleasesServer.URL,
		ReleaseServerPath: "/repos/styrainc/regal/releases/latest",

		CTAURLPrefix: "https://github.com/StyraInc/regal/releases/tag/",

		Debug: true,
	}

	CheckAndWarn(opts, w)

	output := w.String()

	if remoteCalls != 1 {
		t.Errorf("expected 1 remote call, got %d", remoteCalls)
	}

	expectedOutput := `A new version of Regal is available (v0.2.0). You are running v0.1.0.
See https://github.com/StyraInc/regal/releases/tag/v0.2.0 for the latest release.`

	if !strings.Contains(output, expectedOutput) {
		t.Fatalf("expected output to contain\n%s,\ngot\n%s", expectedOutput, output)
	}

	// run the function again and check that the state is loaded from disk
	w = bytes.NewBuffer(nil)
	CheckAndWarn(opts, w)

	if remoteCalls != 1 {
		t.Errorf("expected remote to only be called once, got %d", remoteCalls)
	}

	output = w.String()

	// the same output is expected based on the data on disk
	if !strings.Contains(output, expectedOutput) {
		t.Fatalf("expected output to contain\n%s,\ngot\n%s", expectedOutput, output)
	}

	// update the time to sometime in the future
	opts.CurrentTime = opts.CurrentTime.Add(4 * 24 * time.Hour)

	// run the function again and check that the state is loaded from the remote again
	w = bytes.NewBuffer(nil)
	CheckAndWarn(opts, w)

	if remoteCalls != 2 {
		t.Errorf("expected remote to be called again, got %d", remoteCalls)
	}

	// the same output is expected again
	if !strings.Contains(output, expectedOutput) {
		t.Fatalf("expected output to contain\n%s,\ngot\n%s", expectedOutput, output)
	}

	// if the version is not a semver, then there should be no output
	opts.CurrentVersion = "not-semver"

	w = bytes.NewBuffer(nil)
	CheckAndWarn(opts, w)

	output = w.String()

	if output != "" {
		t.Fatalf("expected no output, got\n%s", output)
	}

	// if the version is greater than the latest version, then there should be no output
	opts.CurrentVersion = "v0.3.0"
	opts.Debug = false

	w = bytes.NewBuffer(nil)
	CheckAndWarn(opts, w)

	output = w.String()

	if output != "" {
		t.Fatalf("expected no output, got\n%s", output)
	}

	// if the version is the same as the latest version, then there should be no output
	opts.CurrentVersion = "v0.2.0"

	w = bytes.NewBuffer(nil)
	CheckAndWarn(opts, w)

	output = w.String()

	if output != "" {
		t.Fatalf("expected no output, got\n%s", output)
	}
}
