package cmd

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
	"github.com/gurbaj5124871/gitcad/core"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the working tree status",
	Long:  "Display the state of the working directory and staging area.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		repoRoot, err := core.FindRepoRoot(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		currentBranch, err := core.GetCurrentBranch(repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		green := color.New(color.FgGreen).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()

		if currentBranch != "" {
			fmt.Printf("On branch %s\n", cyan(currentBranch))
		} else {
			fmt.Println("HEAD detached")
		}

		// Get index
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

		// Calculate staged changes (index vs HEAD)
		var stagedNew, stagedModified, stagedDeleted []string
		for path, entry := range index {
			if committedHash, exists := committedFiles[path]; !exists {
				stagedNew = append(stagedNew, path)
			} else if committedHash != entry.Hash {
				stagedModified = append(stagedModified, path)
			}
		}
		for path := range committedFiles {
			if _, exists := index[path]; !exists {
				stagedDeleted = append(stagedDeleted, path)
			}
		}

		sort.Strings(stagedNew)
		sort.Strings(stagedModified)
		sort.Strings(stagedDeleted)

		if len(stagedNew) > 0 || len(stagedModified) > 0 || len(stagedDeleted) > 0 {
			fmt.Println("\nChanges to be committed:")
			for _, f := range stagedNew {
				fmt.Printf("  %s  %s\n", green("new file:"), green(f))
			}
			for _, f := range stagedModified {
				fmt.Printf("  %s  %s\n", green("modified:"), green(f))
			}
			for _, f := range stagedDeleted {
				fmt.Printf("  %s  %s\n", green("deleted:"), green(f))
			}
		}

		// Calculate unstaged changes (working tree vs index)
		var unstaged, untracked []string
		filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if info.Name() == core.RepoDir {
					return filepath.SkipDir
				}
				return nil
			}

			relPath, _ := filepath.Rel(repoRoot, path)
			if indexEntry, exists := index[relPath]; exists {
				// Check if working tree differs from index
				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				h := sha256.New()
				header := fmt.Sprintf("blob %d\x00", len(content))
				h.Write([]byte(header))
				h.Write(content)
				workingHash := fmt.Sprintf("%x", h.Sum(nil))
				if workingHash != indexEntry.Hash {
					unstaged = append(unstaged, relPath)
				}
			} else {
				untracked = append(untracked, relPath)
			}
			return nil
		})

		sort.Strings(unstaged)
		sort.Strings(untracked)

		if len(unstaged) > 0 {
			fmt.Println("\nChanges not staged for commit:")
			fmt.Println("  (use \"gitcad add <file>...\" to update what will be committed)")
			for _, f := range unstaged {
				fmt.Printf("  %s  %s\n", red("modified:"), red(f))
			}
		}

		if len(untracked) > 0 {
			fmt.Println("\nUntracked files:")
			fmt.Println("  (use \"gitcad add <file>...\" to include in what will be committed)")
			for _, f := range untracked {
				fmt.Printf("  %s\n", red(f))
			}
		}

		if len(stagedNew) == 0 && len(stagedModified) == 0 && len(stagedDeleted) == 0 &&
			len(unstaged) == 0 && len(untracked) == 0 {
			fmt.Println("\nnothing to commit, working tree clean")
		}

		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
