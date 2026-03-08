package core

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ObjectType represents the type of a gitforcad object.
type ObjectType string

const (
	BlobType   ObjectType = "blob"
	TreeType   ObjectType = "tree"
	CommitType ObjectType = "commit"
)

// Object is a generic gitforcad object.
type Object struct {
	Type    ObjectType `json:"type"`
	Content []byte     `json:"content"`
}

// TreeEntry represents an entry in a tree object.
type TreeEntry struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Mode string `json:"mode"` // "file" or "dir"
}

// Tree represents a directory listing.
type Tree struct {
	Entries []TreeEntry `json:"entries"`
}

// Commit represents a commit object.
type Commit struct {
	Tree      string   `json:"tree"`
	Parents   []string `json:"parents"`
	Author    string   `json:"author"`
	Timestamp string   `json:"timestamp"`
	Message   string   `json:"message"`
}

// HashContent computes the SHA-256 hash of content with a type prefix.
func HashContent(objType ObjectType, content []byte) string {
	header := fmt.Sprintf("%s %d\x00", objType, len(content))
	h := sha256.New()
	h.Write([]byte(header))
	h.Write(content)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// WriteObject writes an object to the object store, compressed with zlib.
func WriteObject(repoRoot string, objType ObjectType, content []byte) (string, error) {
	hash := HashContent(objType, content)

	objDir := RepoPath(repoRoot, "objects", hash[:2])
	objPath := filepath.Join(objDir, hash[2:])

	// Already exists
	if _, err := os.Stat(objPath); err == nil {
		return hash, nil
	}

	if err := os.MkdirAll(objDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create object dir: %w", err)
	}

	// Prepare data: type + content
	obj := Object{Type: objType, Content: content}
	data, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal object: %w", err)
	}

	// Compress with zlib
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return "", fmt.Errorf("failed to compress object: %w", err)
	}
	w.Close()

	if err := os.WriteFile(objPath, buf.Bytes(), 0444); err != nil {
		return "", fmt.Errorf("failed to write object: %w", err)
	}

	return hash, nil
}

// ReadObject reads an object from the object store.
func ReadObject(repoRoot, hash string) (*Object, error) {
	if len(hash) < 3 {
		return nil, fmt.Errorf("invalid hash: %s", hash)
	}

	objPath := RepoPath(repoRoot, "objects", hash[:2], hash[2:])

	data, err := os.ReadFile(objPath)
	if err != nil {
		return nil, fmt.Errorf("object not found: %s", hash)
	}

	// Decompress zlib
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decompress object: %w", err)
	}
	defer r.Close()

	decoded, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	var obj Object
	if err := json.Unmarshal(decoded, &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal object: %w", err)
	}

	return &obj, nil
}

// WriteBlob stores a file as a blob object and returns its hash.
func WriteBlob(repoRoot string, filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return WriteObject(repoRoot, BlobType, content)
}

// ReadBlob reads blob content by hash.
func ReadBlob(repoRoot, hash string) ([]byte, error) {
	obj, err := ReadObject(repoRoot, hash)
	if err != nil {
		return nil, err
	}
	if obj.Type != BlobType {
		return nil, fmt.Errorf("object %s is not a blob", hash)
	}
	return obj.Content, nil
}

// WriteTree creates a tree object from entries.
func WriteTree(repoRoot string, entries []TreeEntry) (string, error) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	tree := Tree{Entries: entries}
	data, err := json.Marshal(tree)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tree: %w", err)
	}

	return WriteObject(repoRoot, TreeType, data)
}

// ReadTree reads a tree object by hash.
func ReadTree(repoRoot, hash string) (*Tree, error) {
	obj, err := ReadObject(repoRoot, hash)
	if err != nil {
		return nil, err
	}
	if obj.Type != TreeType {
		return nil, fmt.Errorf("object %s is not a tree", hash)
	}

	var tree Tree
	if err := json.Unmarshal(obj.Content, &tree); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tree: %w", err)
	}
	return &tree, nil
}

// WriteCommit creates a commit object.
func WriteCommit(repoRoot, treeHash string, parents []string, author, message string) (string, error) {
	commit := Commit{
		Tree:      treeHash,
		Parents:   parents,
		Author:    author,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Message:   message,
	}

	data, err := json.Marshal(commit)
	if err != nil {
		return "", fmt.Errorf("failed to marshal commit: %w", err)
	}

	return WriteObject(repoRoot, CommitType, data)
}

// ReadCommit reads a commit object by hash.
func ReadCommit(repoRoot, hash string) (*Commit, error) {
	obj, err := ReadObject(repoRoot, hash)
	if err != nil {
		return nil, err
	}
	if obj.Type != CommitType {
		return nil, fmt.Errorf("object %s is not a commit", hash)
	}

	var commit Commit
	if err := json.Unmarshal(obj.Content, &commit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal commit: %w", err)
	}
	return &commit, nil
}

// GetTreeEntries builds a flat map of path → hash from a tree, recursively.
func GetTreeEntries(repoRoot, treeHash string) (map[string]string, error) {
	result := make(map[string]string)
	return result, getTreeEntriesRecursive(repoRoot, treeHash, "", result)
}

func getTreeEntriesRecursive(repoRoot, treeHash, prefix string, result map[string]string) error {
	tree, err := ReadTree(repoRoot, treeHash)
	if err != nil {
		return err
	}

	for _, entry := range tree.Entries {
		path := entry.Name
		if prefix != "" {
			path = prefix + "/" + entry.Name
		}

		if entry.Mode == "dir" {
			if err := getTreeEntriesRecursive(repoRoot, entry.Hash, path, result); err != nil {
				return err
			}
		} else {
			result[path] = entry.Hash
		}
	}
	return nil
}
