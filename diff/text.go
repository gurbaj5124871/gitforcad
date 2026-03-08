package diff

import (
	"fmt"
	"strings"
)

// DiffText computes a diff for text files with context lines.
func DiffText(filePath string, oldContent, newContent []byte) *DiffResult {
	var oldLines, newLines []string

	if oldContent != nil {
		oldLines = strings.Split(string(oldContent), "\n")
	}
	if newContent != nil {
		newLines = strings.Split(string(newContent), "\n")
	}

	result := &DiffResult{
		FilePath: filePath,
		FileType: "text",
		IsBinary: false,
		IsCAD:    false,
	}

	// Handle new file
	if oldContent == nil {
		result.Summary = fmt.Sprintf("new file: %s (%d lines)", filePath, len(newLines))
		for _, line := range newLines {
			result.Lines = append(result.Lines, DiffLine{Type: "add", Content: line})
		}
		result.Stats.Additions = len(newLines)
		return result
	}

	// Handle deleted file
	if newContent == nil {
		result.Summary = fmt.Sprintf("deleted: %s (%d lines)", filePath, len(oldLines))
		for _, line := range oldLines {
			result.Lines = append(result.Lines, DiffLine{Type: "del", Content: line})
		}
		result.Stats.Deletions = len(oldLines)
		return result
	}

	// Myers-like diff using LCS
	lcs := computeLCS(oldLines, newLines)
	diffLines := buildDiffFromLCS(oldLines, newLines, lcs)

	result.Lines = diffLines
	for _, dl := range diffLines {
		switch dl.Type {
		case "add":
			result.Stats.Additions++
		case "del":
			result.Stats.Deletions++
		}
	}

	result.Summary = fmt.Sprintf("%s: +%d -%d lines", filePath, result.Stats.Additions, result.Stats.Deletions)
	return result
}

// computeLCS uses dynamic programming to find the longest common subsequence.
func computeLCS(a, b []string) [][]int {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	return dp
}

// buildDiffFromLCS builds diff lines from the LCS table.
func buildDiffFromLCS(oldLines, newLines []string, lcs [][]int) []DiffLine {
	var result []DiffLine
	i, j := len(oldLines), len(newLines)

	var stack []DiffLine
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
			stack = append(stack, DiffLine{Type: "ctx", Content: oldLines[i-1]})
			i--
			j--
		} else if j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]) {
			stack = append(stack, DiffLine{Type: "add", Content: newLines[j-1]})
			j--
		} else if i > 0 {
			stack = append(stack, DiffLine{Type: "del", Content: oldLines[i-1]})
			i--
		}
	}

	// Reverse the stack
	for k := len(stack) - 1; k >= 0; k-- {
		result = append(result, stack[k])
	}

	return result
}
