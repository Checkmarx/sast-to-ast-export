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

		// create api client and authenticate
		client, err := internal.NewSASTClient(url, &http.Client{})
		if err != nil {
			panic(err)
		}
		if err2 := client.Authenticate(username, password); err2 != nil {
			panic(err2)
		}

		// start export
		export, err := internal.CreateExport(productName)
		if err != nil {
			panic(err)
		}
		defer export.Clean()

		// fetch users and save to export dir
		usersData, err := client.GetUsersResponseBody()
		if err != nil {
			panic(err)
		}
		if exportErr := export.AddFile(internal.UsersFileName, usersData); exportErr != nil {
			panic(exportErr)
		}

		// fetch teams and save to export dir
		teamsData, err := client.GetTeamsResponseBody()
		if err != nil {
			panic(err)
		}
		if exportErr := export.AddFile(internal.TeamsFileName, teamsData); exportErr != nil {
			panic(exportErr)
		}

		// create export package
		exportFileName, exportErr := export.CreateExportPackage(productName, outputPath)
		if exportErr != nil {
			panic(exportErr)
		}

		fmt.Printf("SAST data exported to %s\n", exportFileName)
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
