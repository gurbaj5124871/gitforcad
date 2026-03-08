package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetHEAD reads the HEAD file and returns either a ref path or a commit hash.
func GetHEAD(repoRoot string) (string, error) {
	headPath := RepoPath(repoRoot, "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return "", fmt.Errorf("failed to read HEAD: %w", err)
	}

	content := strings.TrimSpace(string(data))
	return content, nil
}

// IsDetachedHEAD checks if HEAD points directly to a commit hash.
func IsDetachedHEAD(head string) bool {
	return !strings.HasPrefix(head, "ref: ")
}

// GetCurrentBranch returns the current branch name, or "" if detached.
func GetCurrentBranch(repoRoot string) (string, error) {
	head, err := GetHEAD(repoRoot)
	if err != nil {
		return "", err
	}

	if IsDetachedHEAD(head) {
		return "", nil
	}

	// "ref: refs/heads/main" -> "main"
	ref := strings.TrimPrefix(head, "ref: ")
	return filepath.Base(ref), nil
}

// ResolveHEAD resolves HEAD to a commit hash.
func ResolveHEAD(repoRoot string) (string, error) {
	head, err := GetHEAD(repoRoot)
	if err != nil {
		return "", err
	}

	if IsDetachedHEAD(head) {
		return head, nil
	}

	// Resolve the ref
	ref := strings.TrimPrefix(head, "ref: ")
	return ResolveRef(repoRoot, ref)
}

// ResolveRef reads a reference file and returns the commit hash it points to.
func ResolveRef(repoRoot, ref string) (string, error) {
	refPath := RepoPath(repoRoot, ref)
	data, err := os.ReadFile(refPath)
	if err != nil {
		// Ref doesn't exist yet (e.g., new repo with no commits)
		return "", nil
	}
	return strings.TrimSpace(string(data)), nil
}

// UpdateRef updates a reference to point to a commit hash.
func UpdateRef(repoRoot, ref, hash string) error {
	refPath := RepoPath(repoRoot, ref)
	if err := os.MkdirAll(filepath.Dir(refPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(refPath, []byte(hash+"\n"), 0644)
}

// UpdateHEAD updates HEAD to point to a branch ref.
func UpdateHEAD(repoRoot, branchName string) error {
	headPath := RepoPath(repoRoot, "HEAD")
	content := fmt.Sprintf("ref: refs/heads/%s\n", branchName)
	return os.WriteFile(headPath, []byte(content), 0644)
}

// CreateBranch creates a new branch at the current HEAD commit.
func CreateBranch(repoRoot, name string) error {
	// Check if branch already exists
	refPath := RepoPath(repoRoot, "refs", "heads", name)
	if _, err := os.Stat(refPath); err == nil {
		return fmt.Errorf("branch '%s' already exists", name)
	}

	// Get current HEAD commit
	headHash, err := ResolveHEAD(repoRoot)
	if err != nil {
		return err
	}

	if headHash == "" {
		return fmt.Errorf("cannot create branch: no commits yet")
	}

	return UpdateRef(repoRoot, "refs/heads/"+name, headHash)
}

// ListBranches returns all branch names and the current branch.
func ListBranches(repoRoot string) ([]string, string, error) {
	branchDir := RepoPath(repoRoot, "refs", "heads")
	entries, err := os.ReadDir(branchDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read branches: %w", err)
	}

	currentBranch, err := GetCurrentBranch(repoRoot)
	if err != nil {
		return nil, "", err
	}

	var branches []string
	for _, entry := range entries {
		if !entry.IsDir() {
			branches = append(branches, entry.Name())
		}
	}

	return branches, currentBranch, nil
}

// DeleteBranch deletes a branch (cannot delete current branch).
func DeleteBranch(repoRoot, name string) error {
	currentBranch, err := GetCurrentBranch(repoRoot)
	if err != nil {
		return err
	}
	if name == currentBranch {
		return fmt.Errorf("cannot delete the currently active branch '%s'", name)
	}

	refPath := RepoPath(repoRoot, "refs", "heads", name)
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		return fmt.Errorf("branch '%s' not found", name)
	}

	return os.Remove(refPath)
}

// Checkout switches to the given branch, restoring the working tree.
func Checkout(repoRoot, branchName string) error {
	// Verify branch exists
	refPath := RepoPath(repoRoot, "refs", "heads", branchName)
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		return fmt.Errorf("branch '%s' not found", branchName)
	}

	// Get the commit hash for the target branch
	commitHash, err := ResolveRef(repoRoot, "refs/heads/"+branchName)
	if err != nil {
		return err
	}

	// If there's a commit, restore the working tree
	if commitHash != "" {
		if err := RestoreWorkingTree(repoRoot, commitHash); err != nil {
			return err
		}
	}

	// Update HEAD
	if err := UpdateHEAD(repoRoot, branchName); err != nil {
		return err
	}

	// Update index to match the target branch's tree
	if commitHash != "" {
		commit, err := ReadCommit(repoRoot, commitHash)
		if err != nil {
			return err
		}

		entries, err := GetTreeEntries(repoRoot, commit.Tree)
		if err != nil {
			return err
		}

		index := make(Index)
		for path, hash := range entries {
			index[path] = IndexEntry{Hash: hash, Mode: "file"}
		}
		if err := WriteIndex(repoRoot, index); err != nil {
			return err
		}
	}

	return nil
}

// RestoreWorkingTree restores the working directory to match a commit.
func RestoreWorkingTree(repoRoot, commitHash string) error {
	commit, err := ReadCommit(repoRoot, commitHash)
	if err != nil {
		return err
	}

	// Get all files from the commit's tree
	entries, err := GetTreeEntries(repoRoot, commit.Tree)
	if err != nil {
		return err
	}

	// Clean working directory (remove tracked files not in target)
	err = cleanWorkingDir(repoRoot, entries)
	if err != nil {
		return err
	}

	// Write files from the commit
	for path, hash := range entries {
		content, err := ReadBlob(repoRoot, hash)
		if err != nil {
			return fmt.Errorf("failed to read blob for %s: %w", path, err)
		}

		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	return nil
}

// cleanWorkingDir removes files that are tracked but not in the target tree.
func cleanWorkingDir(repoRoot string, targetEntries map[string]string) error {
	return filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .gitforcad directory
		if info.IsDir() && info.Name() == RepoDir {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}

		// If file is not in target tree, remove it
		if _, exists := targetEntries[relPath]; !exists {
			return os.Remove(path)
		}

		return nil
	})
}
