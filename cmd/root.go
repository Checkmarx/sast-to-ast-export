package cmd

import (
	"ast-sast-export/internal"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	ProductName      = "sast-export"
	UsernameEnvVar   = "SAST_EXPORT_USERNAME"
	PasswordEnvVar   = "SAST_EXPORT_PASSWORD" //nolint:gosec
	ExportFilePrefix = "sast-export"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s [SAST url]", ProductName),
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// process input
		url := args[0]
		outputPath, err := cmd.Flags().GetString("output")
		if err != nil {
			panic(err)
		}
		username := os.Getenv(UsernameEnvVar)
		password := os.Getenv(PasswordEnvVar)
		if username == "" {
			fmt.Printf("please set api username for export with %s env var\n", UsernameEnvVar)
			os.Exit(1)
		}
		if password == "" {
			fmt.Printf("please set api password for export with %s env var\n", PasswordEnvVar)
			os.Exit(1)
		}

		// create api client and authenticate
		client, err := internal.NewSASTClient(url)
		if err != nil {
			panic(err)
		}
		if err := client.Authenticate(username, password); err != nil {
			panic(err)
		}

		// generate export
		export, err := internal.NewExport(client)
		if err != nil {
			panic(err)
		}
		fileName := export.CreateFileName(outputPath, ExportFilePrefix)
		file, err := os.Create(fileName)
		if err != nil {
			panic(err)
		}
		if err := export.WriteToFile(file); err != nil {
			panic(err)
		}
		if err := file.Sync(); err != nil {
			panic(err)
		}
		if err := file.Close(); err != nil {
			panic(err)
		}

		fmt.Printf("SAST data exported to %s", fileName)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().StringP("output", "o", "", "Output path")
}
