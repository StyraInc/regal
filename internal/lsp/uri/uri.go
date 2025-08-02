package uri

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/styrainc/regal/internal/lsp/clients"
)

var (
	drivePattern             = regexp.MustCompile(`^([A-Za-z]):`)
	drivePatternMaybeEncoded = regexp.MustCompile(`^([A-Za-z])(%3[aA]|:)`)
)

// FromPath converts a file path to a URI for a given client.
// Since clients expect URIs to be in a specific format, this function
// will convert the path to the appropriate format for the client.
func FromPath(client clients.Identifier, path string) string {
	path = strings.TrimPrefix(path, "file://")
	path = strings.TrimPrefix(path, "/")

	var driveLetter string
	if matches := drivePattern.FindStringSubmatch(path); len(matches) > 0 {
		driveLetter = matches[1] + ":"
	}

	if driveLetter != "" {
		path = strings.TrimPrefix(path, driveLetter)
	}

	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, part := range parts {
		parts[i] = url.QueryEscape(part)
		parts[i] = strings.ReplaceAll(parts[i], "+", "%20")
	}

	if client == clients.IdentifierVSCode {
		if driveLetter != "" {
			return "file:///" + url.QueryEscape(driveLetter) + strings.Join(parts, "/")
		}

		return "file:///" + strings.Join(parts, "/")
	}

	if driveLetter != "" {
		return "file:///" + driveLetter + strings.Join(parts, "/")
	}

	return "file:///" + strings.Join(parts, "/")
}

// ToPath converts a URI to a file path from a format for a given client.
// Some clients represent URIs differently, and so this function exists to convert
// client URIs into a standard file paths.
func ToPath(client clients.Identifier, uri string) string {
	// if the uri appears to be a URI with a file prefix, then remove the prefix
	path, hadPrefix := strings.CutPrefix(uri, "file://")
	if hadPrefix {
		// if it looks like a URI, then try and decode it
		if decodedPath, err := url.QueryUnescape(path); err == nil {
			path = decodedPath
		}
	}

	if client == clients.IdentifierVSCode {
		path = strings.TrimPrefix(path, "/")
		// handling case for windows when the drive letter is set

		// TODO; never set?
		var driveLetter string

		if matches := drivePatternMaybeEncoded.FindStringSubmatch(path); len(matches) > 1 {
			path = strings.TrimPrefix(path, matches[0])
			path = matches[1] + ":" + strings.TrimPrefix(path, driveLetter)
		} else {
			path = "/" + path
		}
	}

	// Convert path to use system separators
	return filepath.FromSlash(path)
}
