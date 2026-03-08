package diff

import (
	"fmt"
)

// DiffBinary computes a summary diff for binary files.
func DiffBinary(filePath string, oldContent, newContent []byte) *DiffResult {
	result := &DiffResult{
		FilePath: filePath,
		FileType: "binary",
		IsBinary: true,
		IsCAD:    false,
	}

	oldSize := len(oldContent)
	newSize := len(newContent)

	if oldContent == nil {
		result.Summary = fmt.Sprintf("new binary file: %s (%s)", filePath, formatSize(newSize))
		result.Lines = append(result.Lines, DiffLine{
			Type:    "add",
			Content: fmt.Sprintf("Binary file added (%s)", formatSize(newSize)),
		})
		result.Stats.Additions = 1
		return result
	}

	if newContent == nil {
		result.Summary = fmt.Sprintf("deleted binary file: %s (%s)", filePath, formatSize(oldSize))
		result.Lines = append(result.Lines, DiffLine{
			Type:    "del",
			Content: fmt.Sprintf("Binary file deleted (%s)", formatSize(oldSize)),
		})
		result.Stats.Deletions = 1
		return result
	}

	sizeDiff := newSize - oldSize
	var sizeChange string
	if sizeDiff > 0 {
		sizeChange = fmt.Sprintf("+%s", formatSize(sizeDiff))
	} else if sizeDiff < 0 {
		sizeChange = fmt.Sprintf("-%s", formatSize(-sizeDiff))
	} else {
		sizeChange = "same size"
	}

	result.Summary = fmt.Sprintf("binary file changed: %s → %s (%s)",
		formatSize(oldSize), formatSize(newSize), sizeChange)
	result.Lines = append(result.Lines, DiffLine{
		Type:    "del",
		Content: fmt.Sprintf("Old: %s", formatSize(oldSize)),
	})
	result.Lines = append(result.Lines, DiffLine{
		Type:    "add",
		Content: fmt.Sprintf("New: %s (%s)", formatSize(newSize), sizeChange),
	})
	result.Stats.Additions = 1
	result.Stats.Deletions = 1

	return result
}

func formatSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}
