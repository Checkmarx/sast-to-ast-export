package cmd

import (
	"bytes"
	"fmt"
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

		// fetch data
		projects, err := client.GetProjects()
		if err != nil {
			panic(err)
		}

		// generate export
		export := internal.Export{
			FilePrefix: ProductName,
			Data:       internal.ExportData{Projects: projects},
		}
		fileName := export.CreateFileName("")
		buf := bytes.NewBufferString("")
		if err := export.WriteToFile(buf); err != nil {
			panic(err)
		}

		// encrypt
		plaintext := buf.String()
		ciphertext, err := internal.Encrypt(internal.RSAPublicKey, plaintext)
		if err != nil {
			panic(err)
		}

		// write encrypted export to file
		file, err := os.Create(fileName)
		if err != nil {
			panic(err)
		}

		_, err = file.Write([]byte(ciphertext))
		if err := file.Sync(); err != nil {
			panic(err)
		}
		if err := file.Close(); err != nil {
			panic(err)
		}

		fmt.Printf("SAST data exported to %s\n", fileName)
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
