package main

import (
	"os"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/sqlite"
	"github.com/spf13/cobra"
)

// https://github.com/spf13/cobra

// cabby-cli create collection <flags>

// cabby-cli associate -u username -c collectionID -r -w

var (
	cabbyEnv              string
	configPath            string
	collectionAPIRootPath string
	collectionID          string
	collectionTitle       string
	collectionDescription string
	userAdmin             bool
	userName              string
	userPassword          string
)

func cmdCreate() *cobra.Command {
	return &cobra.Command{
		Use:   "create [command/resource]",
		Short: "Create a resource",
		Args:  cobra.MinimumNArgs(1),
	}
}

func cmdDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [command/resource]",
		Short: "Delete a resource",
		Args:  cobra.MinimumNArgs(1),
	}
}

func cmdUpdate() *cobra.Command {
	return &cobra.Command{
		Use:   "update [command/resource]",
		Short: "Update a resource",
		Args:  cobra.MinimumNArgs(1),
	}
}

func dataStoreFromConfig(path, environment string) (cabby.DataStore, error) {
	cs := cabby.Configs{}.Parse(path)
	c := cs[environment]

	return sqlite.NewDataStore(c.DataStore["path"])
}

func main() {
	cabbyEnv = os.Getenv(cabby.CabbyEnvironmentVariable)
	if len(cabbyEnv) == 0 {
		cabbyEnv = cabby.DefaultCabbyEnvironment
	}

	// set up root and subcommands
	var rootCmd = &cobra.Command{Use: "cabby-cli"}
	rootCmd.PersistentFlags().StringVar(&configPath, "config", cabby.CabbyConfigs[cabbyEnv], "path to cabby config file")
	rootCmd.MarkFlagRequired("config")

	cmdCreate := cmdCreate()
	cmdDelete := cmdDelete()
	cmdUpdate := cmdUpdate()
	rootCmd.AddCommand(cmdCreate, cmdDelete, cmdUpdate)

	cmdCreate.AddCommand(cmdCreateCollection(), cmdCreateUser())
	cmdDelete.AddCommand(cmdDeleteCollection(), cmdDeleteUser())
	cmdUpdate.AddCommand(cmdUpdateCollection(), cmdUpdateUser())
	rootCmd.Execute()
}
