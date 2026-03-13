package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/gurbaj5124871/gitcad/core"
	"github.com/spf13/cobra"
)

var deleteBranch string

var branchCmd = &cobra.Command{
	Use:   "branch [name]",
	Short: "List, create, or delete branches",
	Long:  "Without arguments, list all branches. With a name, create a new branch.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		repoRoot, err := core.FindRepoRoot(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Delete branch
		if deleteBranch != "" {
			if err := core.DeleteBranch(repoRoot, deleteBranch); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Deleted branch %s\n", deleteBranch)
			return
		}

		// Create branch
		if len(args) == 1 {
			branchName := args[0]
			if err := core.CreateBranch(repoRoot, branchName); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created branch '%s'\n", branchName)
			return
		}

		// List branches
		branches, currentBranch, err := core.ListBranches(repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		green := color.New(color.FgGreen).SprintFunc()

		for _, b := range branches {
			if b == currentBranch {
				fmt.Printf("* %s\n", green(b))
			} else {
				fmt.Printf("  %s\n", b)
			}
		}
	},
}

func init() {
	branchCmd.Flags().StringVarP(&deleteBranch, "delete", "d", "", "Delete a branch")
	rootCmd.AddCommand(branchCmd)
}
