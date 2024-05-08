package uri

import (
	"net/url"
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
	// if the uri appears to be a URI with a file prefix, then remove the prefix
	path := strings.TrimPrefix(uri, "file://")

	// if it looks like a URI, then try and decode it
	if strings.HasPrefix(uri, "file://") {
		decodedPath, err := url.QueryUnescape(path)
		if err == nil {
			path = decodedPath
		}
	}

	if client == clients.IdentifierVSCode {
		// handling case for windows when the drive letter is set
		if strings.Contains(path, ":") || strings.Contains(path, "%3A") {
			path = strings.Replace(path, "%3A", ":", 1)
			path = strings.TrimPrefix(path, "/")
		}
	}

	// Convert path to use system separators
	path = filepath.FromSlash(path)

	return path
}
