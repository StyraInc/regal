package util

import (
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"strings"

	rio "github.com/styrainc/regal/internal/io"
)

// NullToEmpty returns empty slice if provided slice is nil.
func NullToEmpty[T any](a []T) []T {
	if a == nil {
		return []T{}
	}

	return a
}

// SearchMap searches map for value at provided path.
func SearchMap(object map[string]any, path []string) (any, error) {
	current := object
	traversed := make([]string, 0, len(path))

	for i, p := range path {
		var ok bool
		if i == len(path)-1 {
			value, ok := current[p]
			if ok {
				return value, nil
			}

			return nil, fmt.Errorf("no '%v' attribute at path '%v'", p, strings.Join(traversed, "."))
		}

		if current, ok = current[p].(map[string]any); !ok {
			return nil, fmt.Errorf("no '%v' attribute at path '%v'", p, strings.Join(traversed, "."))
		}

		traversed = append(traversed, p)
	}

	return current, nil
}

// Must0 an error (as commonly returned by Go functions) and panics if the error is not nil.
func Must0(err error) {
	if err != nil {
		panic(err)
	}
}

// Must takes a value and an error (as commonly returned by Go functions) and panics if the error is not nil.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

// Map applies a function to each element of a slice and returns a new slice with the results.
func Map[T, U any](f func(T) U, a []T) []U {
	b := make([]U, len(a))

	for i, v := range a {
		b[i] = f(v)
	}

	return b
}

// FindClosestMatchingRoot finds the closest matching root for a given path.
// If no matching root is found, an empty string is returned.
func FindClosestMatchingRoot(path string, roots []string) string {
	currentLongestSuffix := 0
	longestSuffixIndex := -1

	for i, root := range roots {
		if root == path {
			return path
		}

		if !strings.HasPrefix(path, root) {
			continue
		}

		suffix := strings.TrimPrefix(root, path)

		if len(suffix) > currentLongestSuffix {
			currentLongestSuffix = len(suffix)
			longestSuffixIndex = i
		}
	}

	if longestSuffixIndex == -1 {
		return ""
	}

	return roots[longestSuffixIndex]
}

// FilepathJoiner returns a function that joins provided path with base path.
func FilepathJoiner(base string) func(string) string {
	return func(path string) string {
		return filepath.Join(base, path)
	}
}

// DeleteEmptyDirs will delete empty directories up to the root for a given
// directory.
func DeleteEmptyDirs(dir string) error {
	for {
		// os.Remove will only delete empty directories
		if err := os.Remove(dir); err != nil {
			if os.IsExist(err) {
				break
			}

			if !os.IsNotExist(err) && !os.IsPermission(err) {
				return fmt.Errorf("failed to clean directory %s: %w", dir, err)
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return nil
}

// DirCleanUpPaths will, for a given target file, list all the dirs that would
// be empty if the target file was deleted.
func DirCleanUpPaths(target string, preserve []string) ([]string, error) {
	dirs := make([]string, 0)

	preserveDirs := make(map[string]struct{})

	for _, p := range preserve {
		for {
			preserveDirs[p] = struct{}{}

			p = filepath.Dir(p)

			if p == "." || p == "/" {
				break
			}

			if _, ok := preserveDirs[p]; ok {
				break
			}
		}
	}

	dir := filepath.Dir(target)

	for {
		// check if we reached the preserved dir
		_, ok := preserveDirs[dir]
		if ok {
			break
		}

		parts := strings.Split(dir, rio.PathSeparator)
		if len(parts) == 1 {
			break
		}

		info, err := os.Stat(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to stat directory %s: %w", dir, err)
		}

		if !info.IsDir() {
			return nil, fmt.Errorf("expected directory, got file %s", dir)
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
		}

		empty := true

		for _, file := range files {
			// exclude the target
			abs := filepath.Join(dir, file.Name())
			if abs == target {
				continue
			}

			// exclude any other marked dirs
			if file.IsDir() && len(dirs) > 0 {
				if dirs[len(dirs)-1] == abs {
					continue
				}
			}

			empty = false

			break
		}

		if !empty {
			break
		}

		dirs = append(dirs, dir)

		dir = filepath.Dir(dir)
	}

	return dirs, nil
}

// SafeUintToInt will convert a uint to an int, clamping the result to
// math.MaxInt.
func SafeUintToInt(u uint) int {
	if u > math.MaxInt {
		return math.MaxInt // Clamp to prevent overflow
	}

	return int(u)
}

// FreePort returns a free port to listen on, if none of the preferred ports
// are available then a random free port is returned.
func FreePort(preferred ...int) (port int, err error) {
	listen := func(p int) (int, error) {
		l, err := net.ListenTCP("tcp", &net.TCPAddr{Port: p})
		if err != nil {
			return 0, fmt.Errorf("failed to listen on port %d: %w", p, err)
		}
		defer l.Close()

		addr, ok := l.Addr().(*net.TCPAddr)
		if !ok {
			return 0, errors.New("failed to get port from listener")
		}

		return addr.Port, nil
	}

	for _, p := range preferred {
		if p != 0 {
			if port, err = listen(p); err == nil {
				return port, nil
			}
		}
	}

	// If no preferred port is available, find a random free port using :0
	if port, err = listen(0); err == nil {
		return port, nil
	}

	return 0, fmt.Errorf("failed to find free port: %w", err)
}
