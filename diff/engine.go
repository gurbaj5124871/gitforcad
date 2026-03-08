package diff

import (
	"path/filepath"
	"strings"
)

// DiffLine represents a single line in a diff output.
type DiffLine struct {
	Type    string `json:"type"` // "add", "del", "ctx" (context/unchanged)
	Content string `json:"content"`
}

// DiffResult represents the result of diffing two files.
type DiffResult struct {
	FilePath string     `json:"file_path"`
	FileType string     `json:"file_type"`
	OldHash  string     `json:"old_hash"`
	NewHash  string     `json:"new_hash"`
	Summary  string     `json:"summary"`
	Lines    []DiffLine `json:"lines"`
	Stats    DiffStats  `json:"stats"`
	IsBinary bool       `json:"is_binary"`
	IsCAD    bool       `json:"is_cad"`
}

// DiffStats holds statistics about the diff.
type DiffStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
}

// CAD file extensions we support
var cadExtensions = map[string]string{
	".stl":  "stl",
	".dxf":  "dxf",
	".obj":  "obj",
	".dwg":  "dwg",
	".step": "step",
	".stp":  "step",
}

// DiffFiles computes the diff between two file contents based on file extension.
func DiffFiles(filePath string, oldContent, newContent []byte) *DiffResult {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check if it's a known CAD format
	if cadType, ok := cadExtensions[ext]; ok {
		switch cadType {
		case "stl":
			return DiffSTL(filePath, oldContent, newContent)
		case "dxf":
			return DiffDXF(filePath, oldContent, newContent)
		case "obj":
			return DiffOBJ(filePath, oldContent, newContent)
		case "dwg":
			return DiffDWG(filePath, oldContent, newContent)
		}
	}

	// Check if content appears to be text
	if isTextContent(oldContent) && isTextContent(newContent) {
		return DiffText(filePath, oldContent, newContent)
	}

	// Binary fallback
	return DiffBinary(filePath, oldContent, newContent)
}

// DiffNewFile creates a diff result for a newly added file.
func DiffNewFile(filePath string, content []byte) *DiffResult {
	return DiffFiles(filePath, nil, content)
}

// DiffDeletedFile creates a diff result for a deleted file.
func DiffDeletedFile(filePath string, content []byte) *DiffResult {
	return DiffFiles(filePath, content, nil)
}

// isTextContent checks if content appears to be text (no null bytes in first 8KB).
func isTextContent(content []byte) bool {
	if content == nil {
		return true
	}
	checkLen := len(content)
	if checkLen > 8192 {
		checkLen = 8192
	}
	for i := 0; i < checkLen; i++ {
		if content[i] == 0 {
			return false
		}
	}
	return true
}
