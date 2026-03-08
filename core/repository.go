package core

import (
	"fmt"
	"os"
	"path/filepath"
)

const RepoDir = ".gitforcad"

// InitRepo initializes a new gitforcad repository in the given directory.
func InitRepo(dir string) error {
	repoPath := filepath.Join(dir, RepoDir)

	if _, err := os.Stat(repoPath); err == nil {
		return fmt.Errorf("repository already initialized in %s", dir)
	}

	dirs := []string{
		filepath.Join(repoPath, "objects"),
		filepath.Join(repoPath, "refs", "heads"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// Create HEAD pointing to refs/heads/main
	headPath := filepath.Join(repoPath, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644); err != nil {
		return fmt.Errorf("failed to create HEAD: %w", err)
	}

	// Create empty index
	indexPath := filepath.Join(repoPath, "index")
	if err := os.WriteFile(indexPath, []byte("{}"), 0644); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// FindRepoRoot walks up from the given directory to find a .gitforcad repository.
func FindRepoRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		repoPath := filepath.Join(dir, RepoDir)
		if info, err := os.Stat(repoPath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a gitforcad repository (or any parent up to mount point)")
		}
		dir = parent
	}
}

// RepoPath returns the .gitforcad directory path for the given repo root.
func RepoPath(repoRoot string, parts ...string) string {
	args := append([]string{repoRoot, RepoDir}, parts...)
	return filepath.Join(args...)
}

// IsRepo checks if the given directory is a gitforcad repository.
func IsRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, RepoDir))
	return err == nil
}
