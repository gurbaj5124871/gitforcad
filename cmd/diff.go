package cmd

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
	"github.com/gurbaj5124871/gitcad/core"
	"github.com/gurbaj5124871/gitcad/diff"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [file]",
	Short: "Show changes between working tree and index",
	Long: `Show file changes with red lines for deletions and green lines for additions.
For CAD files (STL, DXF, OBJ), shows structural geometry changes.
For text files, shows line-by-line differences.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		repoRoot, err := core.FindRepoRoot(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		index, err := core.ReadIndex(repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Get committed files from HEAD
		committedFiles := make(map[string]string)
		headHash, _ := core.ResolveHEAD(repoRoot)
		if headHash != "" {
			commit, err := core.ReadCommit(repoRoot, headHash)
			if err == nil {
				committedFiles, _ = core.GetTreeEntries(repoRoot, commit.Tree)
			}
		}

		// Determine which files to diff
		var filesToDiff []string

		if len(args) == 1 {
			// Specific file
			relPath, err := filepath.Rel(repoRoot, filepath.Join(cwd, args[0]))
			if err != nil {
				relPath = args[0]
			}
			filesToDiff = append(filesToDiff, relPath)
		} else {
			// All tracked files - compare working tree with index
			seen := make(map[string]bool)

			// Check indexed files
			for path := range index {
				seen[path] = true
				filesToDiff = append(filesToDiff, path)
			}

			// Check committed files not in index
			for path := range committedFiles {
				if !seen[path] {
					filesToDiff = append(filesToDiff, path)
				}
			}
		}

		sort.Strings(filesToDiff)

		if len(filesToDiff) == 0 {
			fmt.Println("No files to diff.")
			return
		}

		greenFn := color.New(color.FgGreen).SprintFunc()
		redFn := color.New(color.FgRed).SprintFunc()
		cyanFn := color.New(color.FgCyan, color.Bold).SprintFunc()
		yellowFn := color.New(color.FgYellow).SprintFunc()

		hasDiff := false

		for _, filePath := range filesToDiff {
			// Get working tree content
			fullPath := filepath.Join(repoRoot, filePath)
			workingContent, workErr := os.ReadFile(fullPath)

			// Get indexed/committed content
			var oldContent []byte
			if entry, exists := index[filePath]; exists {
				oldContent, _ = core.ReadBlob(repoRoot, entry.Hash)
			} else if hash, exists := committedFiles[filePath]; exists {
				oldContent, _ = core.ReadBlob(repoRoot, hash)
			}

			// Skip if no working tree file and not tracked
			if workErr != nil && oldContent == nil {
				continue
			}

			// Check if actually changed
			if workErr == nil && oldContent != nil {
				h := sha256.New()
				header := fmt.Sprintf("blob %d\x00", len(workingContent))
				h.Write([]byte(header))
				h.Write(workingContent)
				workHash := fmt.Sprintf("%x", h.Sum(nil))

				indexHash := ""
				if entry, exists := index[filePath]; exists {
					indexHash = entry.Hash
				}
				if workHash == indexHash {
					continue
				}
			}

			var result *diff.DiffResult
			if workErr != nil {
				// File deleted from working tree
				result = diff.DiffDeletedFile(filePath, oldContent)
			} else if oldContent == nil {
				// New file
				result = diff.DiffNewFile(filePath, workingContent)
			} else {
				result = diff.DiffFiles(filePath, oldContent, workingContent)
			}

			if result == nil || (result.Stats.Additions == 0 && result.Stats.Deletions == 0) {
				continue
			}

			hasDiff = true

			// Print header
			fmt.Printf("%s %s\n", cyanFn("diff --gitcad"), cyanFn(filePath))

			if result.IsCAD {
				fmt.Printf("%s %s\n", yellowFn("CAD format:"), yellowFn(result.FileType))
			}

			fmt.Printf("%s\n", result.Summary)
			fmt.Println(cyanFn("───────────────────────────────────────────"))

			// Print diff lines with colors
			for _, line := range result.Lines {
				switch line.Type {
				case "add":
					fmt.Printf("%s %s\n", greenFn("+"), greenFn(line.Content))
				case "del":
					fmt.Printf("%s %s\n", redFn("-"), redFn(line.Content))
				case "ctx":
					fmt.Printf("  %s\n", line.Content)
				}
			}

			fmt.Println()
		}

		if !hasDiff {
			fmt.Println("No changes detected.")
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
