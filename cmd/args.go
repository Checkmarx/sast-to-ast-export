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
	args.Export, err = cmd.Flags().GetStringSlice(exportArg)
	if err != nil {
		panic(err)
	}
	args.Debug, err = cmd.Flags().GetBool(debugArg)
	if err != nil {
		panic(err)
	}
	args.OutputFolder, err = cmd.Flags().GetString(outputFolderArg)
	if err != nil {
		panic(err)
	}
	args.InputFolder, err = cmd.Flags().GetString(inputFolderArg)
	if err != nil {
		panic(err)
	}

	args.ProjectsActiveSince, err = cmd.Flags().GetInt(projectsActiveSinceArg)
	if err != nil {
		panic(err)
	}
	args.ProjectsIds, err = cmd.Flags().GetString(projectsIds)
	if err != nil {
		panic(err)
	}
	args.TeamName, err = cmd.Flags().GetString(teamName)
	if err != nil {
		panic(err)
	}
	args.QueryMappingFile, err = cmd.Flags().GetString(queryMapping)
	if err != nil {
		panic(err)
	}

	args.OutputPath, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	return args
}
