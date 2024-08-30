//nolint:errcheck
package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anderseknert/roast/pkg/encoding"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"

	_ "embed"
)

//go:embed update.rego
var updateModule string

const CheckVersionDisableEnvVar = "REGAL_DISABLE_VERSION_CHECK"

type Options struct {
	CurrentVersion string
	CurrentTime    time.Time

	StateDir string

	ReleaseServerHost string
	ReleaseServerPath string

	CTAURLPrefix string

	Debug bool
}

type latestVersionFileContents struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
}

func CheckAndWarn(opts Options, w io.Writer) {
	// this is a shortcut heuristic to avoid and version checking
	// when in dev/test etc.
	if !strings.HasPrefix(opts.CurrentVersion, "v") {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	latestVersion, err := getLatestVersion(ctx, opts)
	if err != nil {
		if opts.Debug {
			w.Write([]byte(err.Error()))
		}

		return
	}

	regoArgs := []func(*rego.Rego){
		rego.Module("update.rego", updateModule),
		rego.Query(`data.update.needs_update`),
		rego.ParsedInput(ast.NewObject(
			ast.Item(ast.StringTerm("current_version"), ast.StringTerm(opts.CurrentVersion)),
			ast.Item(ast.StringTerm("latest_version"), ast.StringTerm(latestVersion)),
		)),
	}

	rs, err := rego.New(regoArgs...).Eval(context.Background())
	if err != nil {
		if opts.Debug {
			w.Write([]byte(err.Error()))
		}

		return
	}

	if !rs.Allowed() {
		if opts.Debug {
			w.Write([]byte("Regal is up to date"))
		}

		return
	}

	ctaURLPrefix := "https://github.com/StyraInc/regal/releases/tag/"
	if opts.CTAURLPrefix != "" {
		ctaURLPrefix = opts.CTAURLPrefix
	}

	ctaURL := ctaURLPrefix + latestVersion

	tmpl := `A new version of Regal is available (%s). You are running %s.
See %s for the latest release.
`

	w.Write([]byte(fmt.Sprintf(tmpl, latestVersion, opts.CurrentVersion, ctaURL)))
}

func getLatestVersion(ctx context.Context, opts Options) (string, error) {
	if opts.StateDir != "" {
		// first, attempt to get the file from previous invocations to save on remote calls
		latestVersionFilePath := filepath.Join(opts.StateDir, "latest_version.json")

		_, err := os.Stat(latestVersionFilePath)
		if err == nil {
			var preExistingState latestVersionFileContents

			file, err := os.Open(latestVersionFilePath)
			if err != nil {
				return "", fmt.Errorf("failed to open file: %w", err)
			}

			json := encoding.JSON()

			err = json.NewDecoder(file).Decode(&preExistingState)
			if err != nil {
				return "", fmt.Errorf("failed to decode existing version state file: %w", err)
			}

			if opts.CurrentTime.Sub(preExistingState.CheckedAt) < 3*24*time.Hour {
				return preExistingState.LatestVersion, nil
			}
		}
	}

	client := http.Client{}

	releaseServerHost := "https://api.github.com"
	if opts.ReleaseServerHost != "" {
		releaseServerHost = strings.TrimSuffix(opts.ReleaseServerHost, "/")

		if !strings.HasPrefix(releaseServerHost, "http") {
			releaseServerHost = "https://" + releaseServerHost
		}
	}

	releaseServerURL, err := url.Parse(releaseServerHost)
	if err != nil {
		return "", fmt.Errorf("failed to parse release server URL: %w", err)
	}

	releaseServerPath := "/repos/styrainc/regal/releases/latest"
	if opts.ReleaseServerPath != "" {
		releaseServerPath = opts.ReleaseServerPath
	}

	releaseServerURL.Path = releaseServerPath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseServerURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var responseData struct {
		TagName string `json:"tag_name"`
	}

	json := encoding.JSON()

	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	stateBs, err := json.MarshalIndent(latestVersionFileContents{
		LatestVersion: responseData.TagName,
		CheckedAt:     opts.CurrentTime,
	}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal state file: %w", err)
	}

	err = os.WriteFile(opts.StateDir+"/latest_version.json", stateBs, 0o600)
	if err != nil {
		return "", fmt.Errorf("failed to write state file: %w", err)
	}

	return responseData.TagName, nil
}
