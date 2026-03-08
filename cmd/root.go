package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitforcad",
	Short: "GitForCAD — Version control for CAD files",
	Long: `GitForCAD is a version control system designed for CAD files.
It provides git-like commands optimized for managing STL, DXF, OBJ, 
and other CAD file formats with intelligent diffing capabilities.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
