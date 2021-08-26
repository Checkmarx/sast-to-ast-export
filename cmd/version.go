package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("Print the version number of %s", productName),
	Long:  fmt.Sprintf(`All software has versions. This is %s's`, productName),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s %s+%s\n", productName, productVersion, productBuild)
	},
}
