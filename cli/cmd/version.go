package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print OmniBase CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("omnibase", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
