package version

import (
	"runtime"
	"strings"
)

// Version stores the version of Regal and is injected at build time.
var Version = ""

// Additional Regal metadata to be injected at build time.
var (
	Vcs       = ""
	Timestamp = ""
	Hostname  = ""
)

// goVersion is the version of Go this was built with
var goVersion = runtime.Version()

// platform is the runtime OS and architecture of this OPA binary
var platform = runtime.GOOS + "/" + runtime.GOARCH

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
			"Commit:        " + vi.Commit,
			"Timestamp:  " + vi.Timestamp,
			"Hostname:   " + vi.Hostname,
		},
		"\n",
	) + "\n"
}

func New() Info {
	return Info{
		Version:   defaultedString(Version, "unknown"),
		GoVersion: goVersion,
		Platform:  platform,
		Commit:    defaultedString(Vcs, "unknown"),
		Timestamp: defaultedString(Timestamp, "unknown"),
		Hostname:  defaultedString(Hostname, "unknown"),
	}
}

func defaultedString(s string, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}
