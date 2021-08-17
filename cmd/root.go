package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sast-export/internal"

	"github.com/spf13/cobra"
)

const (
	ProductName = "cxsast_exporter"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   ProductName,
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

		// create api client and authenticate
		client, err := internal.NewSASTClient(url, &http.Client{})
		if err != nil {
			panic(err)
		}
		if err2 := client.Authenticate(username, password); err2 != nil {
			panic(err2)
		}

		export, err := internal.CreateExport(ProductName)
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(export.TmpDir)

		// fetch users
		usersData, err := client.GetUsersResponseBody()
		if err != nil {
			panic(err)
		}
		if exportErr := export.AddFile(internal.UsersFile, usersData); exportErr != nil {
			panic(exportErr)
		}

		// fetch teams
		teamsData, err := client.GetTeamsResponseBody()
		if err != nil {
			panic(err)
		}
		if exportErr := export.AddFile(internal.TeamsFile, teamsData); exportErr != nil {
			panic(exportErr)
		}

		zipFileName, zipErr := export.CreateZip(ProductName)
		if zipErr != nil {
			panic(zipErr)
		}
		defer os.Remove(zipFileName)

		// encrypt
		plaintext, err := ioutil.ReadFile(zipFileName)
		if err != nil {
			panic(err)
		}

		ciphertext, err := internal.Encrypt(internal.RSAPublicKey, string(plaintext))
		if err != nil {
			panic(err)
		}

		// write encrypted data to file
		exportFileName := internal.CreateFileName(".", ProductName)
		if exportErr := ioutil.WriteFile(exportFileName, []byte(ciphertext), 0600); exportErr != nil {
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
