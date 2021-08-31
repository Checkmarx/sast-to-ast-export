package cmd

import (
	"fmt"
	"net/http"
	"os"
	"sast-export/internal"

	"github.com/spf13/cobra"
)

// productName is defined in Makefile and initialized during build
var productName string

// productVersion is defined in VERSION and initialized during build
var productVersion string

// productBuild is defined in Makefile and initialized during build
var productBuild string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   productName,
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// process input
		url, err := cmd.Flags().GetString("url")
		if err != nil {
			panic(err)
		}
		username, err := cmd.Flags().GetString("user")
		if err != nil {
			panic(err)
		}
		password, err := cmd.Flags().GetString("pass")
		if err != nil {
			panic(err)
		}

		outputPath, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		internal.GetAllData(url, username, password, export, outputPath, productName)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

//nolint:gochecknoinits
func init() {
	rootCmd.Flags().StringP("user", "", "", "SAST admin username")
	rootCmd.Flags().StringP("pass", "", "", "SAST admin password")
	rootCmd.Flags().StringP("url", "", "", "SAST url")
	if err := rootCmd.MarkFlagRequired("user"); err != nil {
		panic(err)
	}
	if err := rootCmd.MarkFlagRequired("pass"); err != nil {
		panic(err)
	}
	if err := rootCmd.MarkFlagRequired("url"); err != nil {
		panic(err)
	}
}
