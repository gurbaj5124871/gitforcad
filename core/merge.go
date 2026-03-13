package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// MergeResult describes the outcome of a merge.
type MergeResult struct {
	Type       string   // "fast-forward", "merge", "conflict", "up-to-date"
	CommitHash string   // resulting commit hash
	Conflicts  []string // conflicting file paths (if any)
	Message    string   // human-readable message
}

// Merge merges the given branch into the current branch.
func Merge(repoRoot, branchName string) (*MergeResult, error) {
	currentBranch, err := GetCurrentBranch(repoRoot)
	if err != nil {
		return nil, err
	}
	if currentBranch == "" {
		return nil, fmt.Errorf("cannot merge in detached HEAD state")
	}
	if currentBranch == branchName {
		return nil, fmt.Errorf("cannot merge branch '%s' into itself", branchName)
	}

	currentHash, err := ResolveHEAD(repoRoot)
	if err != nil {
		return nil, err
	}

	targetHash, err := ResolveRef(repoRoot, "refs/heads/"+branchName)
	if err != nil {
		return nil, err
	}
	if targetHash == "" {
		return nil, fmt.Errorf("branch '%s' not found or has no commits", branchName)
	}

	if currentHash == targetHash {
		return &MergeResult{Type: "up-to-date", Message: "Already up to date."}, nil
	}

	// Check fast-forward
	if currentHash == "" || isAncestor(repoRoot, currentHash, targetHash) {
		if err := UpdateRef(repoRoot, "refs/heads/"+currentBranch, targetHash); err != nil {
			return nil, err
		}
		if err := RestoreWorkingTree(repoRoot, targetHash); err != nil {
			return nil, err
		}
		if err := rebuildIndexFromCommit(repoRoot, targetHash); err != nil {
			return nil, err
		}
		return &MergeResult{
			Type: "fast-forward", CommitHash: targetHash,
			Message: fmt.Sprintf("Fast-forward merge to %s", targetHash[:8]),
		}, nil
	}

	// Already up-to-date (target is ancestor of current)
	if isAncestor(repoRoot, targetHash, currentHash) {
		return &MergeResult{Type: "up-to-date", Message: "Already up to date."}, nil
	}

	// Three-way merge
	currentCommit, err := ReadCommit(repoRoot, currentHash)
	if err != nil {
		return nil, err
	}
	targetCommit, err := ReadCommit(repoRoot, targetHash)
	if err != nil {
		return nil, err
	}

	currentFiles, err := GetTreeEntries(repoRoot, currentCommit.Tree)
	if err != nil {
		return nil, err
	}
	targetFiles, err := GetTreeEntries(repoRoot, targetCommit.Tree)
	if err != nil {
		return nil, err
	}

	mergedIndex := make(Index)
	var conflicts []string

	for path, hash := range currentFiles {
		mergedIndex[path] = IndexEntry{Hash: hash, Mode: "file"}
	}

	for path, targetFileHash := range targetFiles {
		currentFileHash, existsInCurrent := currentFiles[path]
		if !existsInCurrent {
			mergedIndex[path] = IndexEntry{Hash: targetFileHash, Mode: "file"}
			content, err := ReadBlob(repoRoot, targetFileHash)
			if err != nil {
				return nil, err
			}
			fullPath := filepath.Join(repoRoot, path)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return nil, err
			}
			if err := os.WriteFile(fullPath, content, 0644); err != nil {
				return nil, err
			}
		} else if currentFileHash != targetFileHash {
			conflicts = append(conflicts, path)
		}
	}

	if len(conflicts) > 0 {
		return &MergeResult{
			Type: "conflict", Conflicts: conflicts,
			Message: fmt.Sprintf("Merge conflict in %d file(s)", len(conflicts)),
		}, nil
	}

	treeHash, err := BuildTreeFromIndex(repoRoot, mergedIndex)
	if err != nil {
		return nil, err
	}

	commitHash, err := WriteCommit(repoRoot, treeHash, []string{currentHash, targetHash},
		"gitcad", fmt.Sprintf("Merge branch '%s' into %s", branchName, currentBranch))
	if err != nil {
		return nil, err
	}

	if err := UpdateRef(repoRoot, "refs/heads/"+currentBranch, commitHash); err != nil {
		return nil, err
	}
	if err := WriteIndex(repoRoot, mergedIndex); err != nil {
		return nil, err
	}

	return &MergeResult{
		Type: "merge", CommitHash: commitHash,
		Message: fmt.Sprintf("Merge branch '%s' into %s", branchName, currentBranch),
	}, nil
}

func rebuildIndexFromCommit(repoRoot, commitHash string) error {
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
	return WriteIndex(repoRoot, index)
}

func isAncestor(repoRoot, ancestor, descendant string) bool {
	visited := make(map[string]bool)
	queue := []string{descendant}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == ancestor {
			return true
		}
		if visited[current] || current == "" {
			continue
		}
		visited[current] = true

		commit, err := ReadCommit(repoRoot, current)
		if err != nil {
			continue
		}
		queue = append(queue, commit.Parents...)
	}
	return false
}
