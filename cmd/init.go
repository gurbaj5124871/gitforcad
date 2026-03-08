package cmd

import (
	"fmt"
	"os"

	"github.com/gurbaj5124871/gitforcad/core"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new gitforcad repository",
	Long:  "Create a new gitforcad repository in the current directory with the default 'main' branch.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := core.InitRepo(cwd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Initialized empty gitforcad repository in %s/.gitforcad/\n", cwd)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
