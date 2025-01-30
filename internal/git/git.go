package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"

	"github.com/open-policy-agent/opa/v1/util"
)

// FindGitRepo checks if the given directories belong to the same Git repository.
// It returns the root of the Git repository if all directories share the same repository.
// If the directories belong to different repositories it retruns an error.
func FindGitRepo(dirs ...string) (string, error) {
	if len(dirs) == 0 {
		return "", errors.New("no directories provided")
	}

	var commonRepoPath string

	for _, dir := range dirs {
		repoPath, err := findRepoPath(dir)
		if err != nil {
			return "", err
		}

		if commonRepoPath == "" {
			commonRepoPath = repoPath
		} else if repoPath != commonRepoPath {
			return "", errors.New("directories belong to different Git repositories")
		}
	}

	return commonRepoPath, nil
}

// findRepoPath traverses the directory upwards to find the Git repository's root path.
func findRepoPath(dir string) (string, error) {
	for {
		gitPath := filepath.Join(dir, ".git")

		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to check .git directory: %w", err)
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			break
		}

		dir = parent
	}

	return "", nil
}

// GetChangedFiles will return a list of files that have been changed in the
// repository. This should only be called when required, i.e. a .git directory
// has already been detected.
func GetChangedFiles(dir string) ([]string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	return util.Keys(status), nil
}
