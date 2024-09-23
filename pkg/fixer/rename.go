package fixer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// renameCandidate takes a filename and produces a new name with an incremented
// numeric suffix. It correctly handles test files by inserting the increment
// before the "_test" suffix and preserves the original directory.
func renameCandidate(oldName string) string {
	dir := filepath.Dir(oldName)
	baseWithExt := filepath.Base(oldName)

	ext := filepath.Ext(baseWithExt)
	base := strings.TrimSuffix(baseWithExt, ext)

	suffix := ""
	if strings.HasSuffix(base, "_test") {
		suffix = "_test"
		base = strings.TrimSuffix(base, "_test")
	}

	re := regexp.MustCompile(`^(.*)_(\d+)$`)
	matches := re.FindStringSubmatch(base)

	if len(matches) == 3 {
		baseName := matches[1]
		numStr := matches[2]
		num, _ := strconv.Atoi(numStr)
		num++
		base = fmt.Sprintf("%s_%d", baseName, num)
	} else {
		base += "_1"
	}

	newBase := base + suffix + ext
	newName := filepath.Join(dir, newBase)

	return newName
}
