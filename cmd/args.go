package cmd

import (
	"os"

	"github.com/checkmarxDev/ast-sast-export/internal"

	"github.com/spf13/cobra"
)

func GetArgs(cmd *cobra.Command, productName string) internal.Args {
	args := internal.Args{}
	args.ProductName = productName
	var err error
	// process input
	args.URL, err = cmd.Flags().GetString(urlArg)
	if err != nil {
		panic(err)
	}
	args.Username, err = cmd.Flags().GetString(userArg)
	if err != nil {
		panic(err)
	}
	args.Password, err = cmd.Flags().GetString(passArg)
	if err != nil {
		panic(err)
	}
	args.Export, err = cmd.Flags().GetString(exportArg)
	if err != nil {
		panic(err)
	}
	args.Debug, err = cmd.Flags().GetBool(debugArg)
	if err != nil {
		panic(err)
	}
	args.ResultsProjectActiveSince, err = cmd.Flags().GetInt(resultsProjectActiveSinceArg)
	if err != nil {
		panic(err)
	}

	args.OutputPath, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	return args
}
