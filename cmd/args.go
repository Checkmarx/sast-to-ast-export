package cmd

import (
	"os"
	"sast-export/internal"

	"github.com/spf13/cobra"
)

func GetArgs(cmd *cobra.Command, productName string) {
	args := internal.Args{}
	args.ProductName = productName
	var err error
	// process input
	args.Url, err = cmd.Flags().GetString("url")
	if err != nil {
		panic(err)
	}
	args.Username, err = cmd.Flags().GetString("user")
	if err != nil {
		panic(err)
	}
	args.Password, err = cmd.Flags().GetString("pass")
	if err != nil {
		panic(err)
	}
	args.Export, err = cmd.Flags().GetString("export")
	if err != nil {
		panic(err)
	}
	args.Debug, err = cmd.Flags().GetBool("debug")
	if err != nil {
		panic(err)
	}

	args.OutputPath, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	internal.RunExport(args)
}
