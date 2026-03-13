package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/gurbaj5124871/gitcad/core"
	"github.com/spf13/cobra"
)

var logCount int

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show commit history",
	Long:  "Display the commit log for the current branch.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		repoRoot, err := core.FindRepoRoot(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		headHash, err := core.ResolveHEAD(repoRoot)
		if err != nil || headHash == "" {
			fmt.Println("No commits yet.")
			return
		}

		yellow := color.New(color.FgYellow).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()

		currentBranch, _ := core.GetCurrentBranch(repoRoot)

		count := 0
		hash := headHash
		for hash != "" && (logCount <= 0 || count < logCount) {
			commit, err := core.ReadCommit(repoRoot, hash)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading commit: %v\n", err)
				break
			}

			// Header
			branchLabel := ""
			if count == 0 && currentBranch != "" {
				branchLabel = fmt.Sprintf(" -> %s", cyan(currentBranch))
			}
			fmt.Printf("%s %s%s\n", yellow("commit"), yellow(hash[:12]), branchLabel)
			fmt.Printf("Author: %s\n", commit.Author)
			fmt.Printf("Date:   %s\n", commit.Timestamp)
			fmt.Printf("\n    %s\n\n", commit.Message)

			// Follow first parent
			if len(commit.Parents) > 0 {
				hash = commit.Parents[0]
			} else {
				hash = ""
			}
			count++
		}
	},
}

func init() {
	logCmd.Flags().IntVarP(&logCount, "number", "n", 0, "Limit the number of commits shown")
	rootCmd.AddCommand(logCmd)
}
