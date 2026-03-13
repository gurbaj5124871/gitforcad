package cmd

import (
	"fmt"
	"os"

	"github.com/gurbaj5124871/gitcad/core"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <file>...",
	Short: "Stage files for the next commit",
	Long:  "Add file contents to the staging area. Use '.' to add all files.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		repoRoot, err := core.FindRepoRoot(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := core.AddToIndex(repoRoot, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		for _, f := range args {
			fmt.Printf("  added: %s\n", f)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
