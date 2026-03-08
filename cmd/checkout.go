package cmd

import (
	"fmt"
	"os"

	"github.com/gurbaj5124871/gitforcad/core"
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout <branch>",
	Short: "Switch branches",
	Long:  "Switch to the specified branch, updating the working directory.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branchName := args[0]

		cwd, _ := os.Getwd()
		repoRoot, err := core.FindRepoRoot(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		currentBranch, _ := core.GetCurrentBranch(repoRoot)
		if currentBranch == branchName {
			fmt.Printf("Already on '%s'\n", branchName)
			return
		}

		if err := core.Checkout(repoRoot, branchName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Switched to branch '%s'\n", branchName)
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
