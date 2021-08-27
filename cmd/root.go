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
		debug, err := cmd.Flags().GetBool("debug")
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
		if !debug {
			defer export.Clean()
		}

		//todo: if USERS cmd selected, move to function
		if true {
			usersData, err := client.GetUsers()
			if err != nil {
				panic(err)
			}
			if exportErr := export.AddFile(internal.UsersFileName, usersData); exportErr != nil {
				panic(exportErr)
			}

			rolesData, err := client.GetRoles()
			if err != nil {
				panic(err)
			}
			if exportErr := export.AddFile(internal.RolesFileName, rolesData); exportErr != nil {
				panic(exportErr)
			}

			ldapRolesData, err := client.GetLdapRoleMappings()
			if err != nil {
				panic(err)
			}
			if exportErr := export.AddFile(internal.LdapRoleMappingsFileName, ldapRolesData); exportErr != nil {
				panic(exportErr)
			}

			samlRolesData, err := client.GetSamlRoleMappings()
			if err != nil {
				panic(err)
			}
			if exportErr := export.AddFile(internal.SamlRoleMappingsFileName, samlRolesData); exportErr != nil {
				panic(exportErr)
			}
		}

		//todo: if USERS | TEAMS selected, refactor to function
		if true {
			ldapServersData, err := client.GetLdapServers()
			if err != nil {
				panic(err)
			}
			if exportErr := export.AddFile(internal.LdapServersFileName, ldapServersData); exportErr != nil {
				panic(exportErr)
			}

			samlIdpsData, err := client.GetSamlIdentityProviders()
			if err != nil {
				panic(err)
			}
			if exportErr := export.AddFile(internal.SamlIdpFileName, samlIdpsData); exportErr != nil {
				panic(exportErr)
			}
		}

		// fetch teams and save to export dir
		//todo: if TEAMS selected, refactor to function
		if true {
			teamsData, err := client.GetTeams()
			if err != nil {
				panic(err)
			}
			if exportErr := export.AddFile(internal.TeamsFileName, teamsData); exportErr != nil {
				panic(exportErr)
			}

			ldapTeamsData, err := client.GetLdapTeamMappings()
			if err != nil {
				panic(err)
			}
			if exportErr := export.AddFile(internal.LdapTeamMappingsFileName, ldapTeamsData); exportErr != nil {
				panic(exportErr)
			}

			samlTeamsData, err := client.GetSamlTeamMappings()
			if err != nil {
				panic(err)
			}
			if exportErr := export.AddFile(internal.SamlTeamMappingsFileName, samlTeamsData); exportErr != nil {
				panic(exportErr)
			}
		}

		// fetch results and save to export dir
		//todo: RESULTS selected, implement....

		// create export package
		if !debug {
			exportFileName, exportErr := export.CreateExportPackage(productName, outputPath)
			if exportErr != nil {
				panic(exportErr)
			}
			fmt.Printf("SAST data exported to %s\n", exportFileName)
		} else {
			fmt.Printf("Debug mode: SAST data exported to %s\n", export.TmpDir)
		}
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
}
