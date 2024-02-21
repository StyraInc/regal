package uri

import (
	"path/filepath"
	"strings"

	"github.com/styrainc/regal/internal/lsp/clients"
)

// FromPath converts a file path to a URI for a given client.
// Since clients expect URIs to be in a specific format, this function
// will convert the path to the appropriate format for the client.
func FromPath(client clients.Identifier, path string) string {
	path = strings.TrimPrefix(path, "file://")
	path = strings.TrimPrefix(path, "/")

	if client == clients.IdentifierVSCode {
		// Convert Windows path separators to Unix separators
		path = filepath.ToSlash(path)

		// If the path is a Windows path, the colon after the drive letter needs to be
		// percent-encoded.
		if parts := strings.Split(path, ":"); len(parts) > 1 {
			path = parts[0] + "%3A" + parts[1]
		}
	}

	return "file://" + "/" + path
}

// ToPath converts a URI to a file path from a format for a given client.
// Some clients represent URIs differently, and so this function exists to convert
// client URIs into a standard file paths.
func ToPath(client clients.Identifier, uri string) string {
	path := strings.TrimPrefix(uri, "file://")

	if client == clients.IdentifierVSCode {
		if strings.Contains(path, ":") || strings.Contains(path, "%3A") {
			path = strings.Replace(path, "%3A", ":", 1)
			path = strings.TrimPrefix(path, "/")
		}
	}

	return path
}
