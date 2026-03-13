package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IndexEntry represents a staged file in the index.
type IndexEntry struct {
	Hash string `json:"hash"`
	Mode string `json:"mode"`
}

// Index represents the staging area.
type Index map[string]IndexEntry

// ReadIndex reads the current index from disk.
func ReadIndex(repoRoot string) (Index, error) {
	indexPath := RepoPath(repoRoot, "index")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	return index, nil
}

// WriteIndex writes the index to disk.
func WriteIndex(repoRoot string, index Index) error {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	indexPath := RepoPath(repoRoot, "index")
	return os.WriteFile(indexPath, data, 0644)
}

// AddToIndex stages a file by hashing it and updating the index.
func AddToIndex(repoRoot string, filePaths []string) error {
	index, err := ReadIndex(repoRoot)
	if err != nil {
		return err
	}

	for _, fp := range filePaths {
		// Get absolute path of the file
		absPath, err := filepath.Abs(fp)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", fp, err)
		}

		// Check file exists
		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("file not found: %s", fp)
		}

		if info.IsDir() {
			// Recursively add directory contents
			err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					// Skip .gitcad directory
					if info.Name() == RepoDir {
						return filepath.SkipDir
					}
					return nil
				}

				relPath, err := filepath.Rel(repoRoot, path)
				if err != nil {
					return err
				}

				hash, err := WriteBlob(repoRoot, path)
				if err != nil {
					return err
				}

				index[relPath] = IndexEntry{Hash: hash, Mode: "file"}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to add directory %s: %w", fp, err)
			}
		} else {
			// Get relative path from repo root
			relPath, err := filepath.Rel(repoRoot, absPath)
			if err != nil {
				return fmt.Errorf("failed to compute relative path: %w", err)
			}

			hash, err := WriteBlob(repoRoot, absPath)
			if err != nil {
				return err
			}

			index[relPath] = IndexEntry{Hash: hash, Mode: "file"}
		}
	}

	return WriteIndex(repoRoot, index)
}

// BuildTreeFromIndex creates a tree object from the current index.
func BuildTreeFromIndex(repoRoot string, index Index) (string, error) {
	// Build nested directory structure
	type dirNode struct {
		files   []TreeEntry
		subdirs map[string]*dirNode
	}

	root := &dirNode{subdirs: make(map[string]*dirNode)}

	for path, entry := range index {
		parts := strings.Split(filepath.ToSlash(path), "/")
		current := root

		// Navigate/create subdirectories
		for _, dir := range parts[:len(parts)-1] {
			if _, ok := current.subdirs[dir]; !ok {
				current.subdirs[dir] = &dirNode{subdirs: make(map[string]*dirNode)}
			}
			current = current.subdirs[dir]
		}

		// Add file entry
		fileName := parts[len(parts)-1]
		current.files = append(current.files, TreeEntry{
			Name: fileName,
			Hash: entry.Hash,
			Mode: entry.Mode,
		})
	}

	// Recursively build tree objects
	var buildTree func(node *dirNode) (string, error)
	buildTree = func(node *dirNode) (string, error) {
		var entries []TreeEntry

		// Add file entries
		entries = append(entries, node.files...)

		// Add subdirectory entries
		for name, subdir := range node.subdirs {
			subHash, err := buildTree(subdir)
			if err != nil {
				return "", err
			}
			entries = append(entries, TreeEntry{
				Name: name,
				Hash: subHash,
				Mode: "dir",
			})
		}

		return WriteTree(repoRoot, entries)
	}

	return buildTree(root)
}

// ClearIndex resets the index to empty.
func ClearIndex(repoRoot string) error {
	return WriteIndex(repoRoot, make(Index))
}
