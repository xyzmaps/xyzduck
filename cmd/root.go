package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"org.xyzmaps.xyzduck/src/version"
)

var rootCmd = &cobra.Command{
	Use:   "xyzduck",
	Short: "xyzduck - A CLI tool",
	Long:  `xyzduck is a CLI application for XYZ Maps`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("xyzduck")
		fmt.Println("Run 'xyzduck --help' for usage information")
	},
}

var versionFlag bool

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version information")

	// Handle version flag
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Println(version.GetFullVersion())
			os.Exit(0)
		}
	}
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
