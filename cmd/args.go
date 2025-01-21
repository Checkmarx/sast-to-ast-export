package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/checkmarxDev/ast-sast-export/internal"
	"github.com/rs/zerolog/log"

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
	args.NestedTeams, err = cmd.Flags().GetBool(nestedTeams)
	if err != nil {
		panic(err)
	}
	args.IsDefaultProjectActiveSince = args.ProjectsActiveSince == emptyProjectsActiveSince
	if args.IsDefaultProjectActiveSince {
		args.ProjectsActiveSince = projectsActiveSinceDefaultValue
	}
	args.OutputPath, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	args.SimIDVersion, err = cmd.Flags().GetInt(simIDVersionArg)
	if err != nil {
		panic(err)
	}
	args.ExcludeFile, err = cmd.Flags().GetString(excludeFileArg)
	if err != nil {
		panic(err)
	}
	if args.ExcludeFile != "" {
		if _, err := os.Stat(args.ExcludeFile); os.IsNotExist(err) {
			log.Fatal().Msgf("Exclude file '%s' does not exist", args.ExcludeFile)
		}
		if filepath.Ext(args.ExcludeFile) != ".txt" {
			log.Fatal().Msgf("Exclude file '%s' must be a .txt file", args.ExcludeFile)
		}
		fileContent, err := os.ReadFile(args.ExcludeFile)
		if err != nil {
			log.Fatal().Err(err).Msgf("Error reading exclude file: %s", args.ExcludeFile)
		}
		excludePaths := strings.Split(string(fileContent), "\n")
		for i := range excludePaths {
			excludePaths[i] = strings.TrimSpace(excludePaths[i])
		}
	}
	return args
}
