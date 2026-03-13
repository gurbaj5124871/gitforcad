package cmd

import (
	"fmt"
	"os"

	"github.com/gurbaj5124871/gitcad/core"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge <branch>",
	Short: "Merge a branch into the current branch",
	Long:  "Join the specified branch history into the current branch.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branchName := args[0]

		cwd, _ := os.Getwd()
		repoRoot, err := core.FindRepoRoot(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		result, err := core.Merge(repoRoot, branchName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result.Message)

		if result.Type == "conflict" {
			fmt.Println("\nConflicting files:")
			for _, f := range result.Conflicts {
				fmt.Printf("  %s\n", f)
			}
			fmt.Println("\nResolve conflicts and commit the result.")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}
