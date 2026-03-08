package cmd

import (
	"fmt"
	"os"
	"os/user"

	"github.com/gurbaj5124871/gitforcad/core"
	"github.com/spf13/cobra"
)

var commitMessage string

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Record changes to the repository",
	Long:  "Create a new commit with the staged changes.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if commitMessage == "" {
			fmt.Fprintln(os.Stderr, "Error: commit message required (-m flag)")
			os.Exit(1)
		}

		cwd, _ := os.Getwd()
		repoRoot, err := core.FindRepoRoot(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Read current index
		index, err := core.ReadIndex(repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(index) == 0 {
			fmt.Fprintln(os.Stderr, "Error: nothing to commit (empty staging area)")
			os.Exit(1)
		}

		// Build tree from index
		treeHash, err := core.BuildTreeFromIndex(repoRoot, index)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building tree: %v\n", err)
			os.Exit(1)
		}

		// Get parent commit
		var parents []string
		headHash, err := core.ResolveHEAD(repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving HEAD: %v\n", err)
			os.Exit(1)
		}
		if headHash != "" {
			parents = []string{headHash}
		}

		// Get author
		author := "unknown"
		if u, err := user.Current(); err == nil {
			author = u.Username
		}

		// Create commit
		commitHash, err := core.WriteCommit(repoRoot, treeHash, parents, author, commitMessage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Update current branch ref
		currentBranch, err := core.GetCurrentBranch(repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if currentBranch != "" {
			if err := core.UpdateRef(repoRoot, "refs/heads/"+currentBranch, commitHash); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("[%s %s] %s\n", currentBranch, commitHash[:8], commitMessage)
		fmt.Printf(" %d file(s) committed\n", len(index))
	},
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message")
	rootCmd.AddCommand(commitCmd)
}
