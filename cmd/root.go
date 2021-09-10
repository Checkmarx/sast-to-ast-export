package cmd

import (
	"os"

	"github.com/checkmarxDev/ast-sast-export/internal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/checkmarxDev/ast-observability-library/pkg/aol"
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
		allArgs := GetArgs(cmd, productName)
		internal.RunExport(allArgs)
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
	rootCmd.Flags().StringP("export", "", "", "SAST [optional] export options --export users,results,teams, all if nothing defined")
	rootCmd.Flags().IntP("results-project-active-since", "", 180, "SAST [optional] custom results project active since (days) - 180 if nothing defined")
	rootCmd.Flags().Bool("debug", false, "Activate debug mode")
	if err := rootCmd.MarkFlagRequired("user"); err != nil {
		panic(err)
	}
	if err := rootCmd.MarkFlagRequired("pass"); err != nil {
		panic(err)
	}
	if err := rootCmd.MarkFlagRequired("url"); err != nil {
		panic(err)
	}
	if err := rootCmd.MarkFlagCustom("export", "users,results,teams"); err != nil {
		panic(err)
	}
	if err := rootCmd.MarkFlagCustom("results-project-active-since", "SAST custom results project active since (days) 180 if nothing defined"); err != nil {
		panic(err)
	}

	aolErr := aol.Init(productName, "", "trace", "")
	if aolErr != nil {
		panic(aolErr)
	}

	logFileWriter, err := os.Create("test.log")
	if err != nil {
		panic(err)
	}
	defer logFileWriter.Close()

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	levelWriter := internal.NewMultiLevelWriter(false, zerolog.InfoLevel, consoleWriter, logFileWriter)
	log.Logger = log.Logger.Output(&levelWriter)

	log.Info().Msg("Inited")
}
