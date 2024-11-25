//nolint:gochecknoglobals
package version

import (
	"runtime"
	"strings"
)

// Version stores the version of Regal and is injected at build time.
var Version = ""

// Additional Regal metadata to be injected at build time.
var (
	Commit    = ""
	Timestamp = ""
	Hostname  = ""
)

// goVersion is the version of Go this was built with.
var goVersion = runtime.Version()

// platform is the runtime OS and architecture of this OPA binary.
const platform = runtime.GOOS + "/" + runtime.GOARCH

// Info wraps the various version metadata values and provides a means of marshalling as JSON or pretty string.
type Info struct {
	Version string `json:"version"`

	GoVersion string `json:"go_version"`

	Platform string `json:"platform"`

	Commit    string `json:"commit"`
	Timestamp string `json:"timestamp"`
	Hostname  string `json:"hostname"`
}

func (vi Info) String() string {
	return strings.Join(
		[]string{
			"Version:    " + vi.Version,
			"Go Version: " + vi.GoVersion,
			"Platform:   " + vi.Platform,
			"Commit:     " + vi.Commit,
			"Timestamp:  " + vi.Timestamp,
			"Hostname:   " + vi.Hostname,
		},
		"\n",
	) + "\n"
}

func New() Info {
	return Info{
		Version:   unknownVersionString(Version),
		GoVersion: goVersion,
		Platform:  platform,
		Commit:    unknownString(Commit),
		Timestamp: unknownString(Timestamp),
		Hostname:  unknownString(Hostname),
	}
}

func unknownString(s string) string {
	if s == "" {
		return "unknown"
	}

	return s
}

func unknownVersionString(s string) string {
	if s == "" {
		return "v.29.3"
	}

	return s
}
