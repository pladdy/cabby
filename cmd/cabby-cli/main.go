package main

import (
	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/sqlite"
	"github.com/spf13/cobra"
)

var (
	apiRootDescription     string
	apiRootPath            string
	apiRootTitle           string
	apiRootVersions        string
	cabbyEnv               string
	configPath             string
	collectionID           string
	collectionTitle        string
	collectionDescription  string
	discoveryContact       string
	discoveryDefault       string
	discoveryDescription   string
	discoveryTitle         string
	maxContentLength       int64
	userAdmin              bool
	userCollectionCanRead  bool
	userCollectionCanWrite bool
	userName               string
	userPassword           string
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

func dataStoreFromConfig(path string) (cabby.DataStore, error) {
	config := cabby.Config{}.Parse(path)
	return sqlite.NewDataStore(config.DataStore["path"])
}

func main() {
	// set up root and subcommands
	var rootCmd = &cobra.Command{Use: "cabby-cli"}
	rootCmd.PersistentFlags().StringVar(&configPath, "config", cabby.DefaultProductionConfig, "path to cabby config file")
	rootCmd.MarkFlagRequired("config")

	cmdCreate := cmdCreate()
	cmdDelete := cmdDelete()
	cmdUpdate := cmdUpdate()
	rootCmd.AddCommand(cmdCreate, cmdDelete, cmdUpdate)

	cmdCreate.AddCommand(
		cmdCreateAPIRoot(),
		cmdCreateCollection(),
		cmdCreateDiscovery(),
		cmdCreateUser(),
		cmdCreateUserCollection())

	cmdDelete.AddCommand(
		cmdDeleteAPIRoot(),
		cmdDeleteCollection(),
		cmdDeleteDiscovery(),
		cmdDeleteUser(),
		cmdDeleteUserCollection())

	cmdUpdate.AddCommand(
		cmdUpdateAPIRoot(),
		cmdUpdateCollection(),
		cmdUpdateDiscovery(),
		cmdUpdateUser(),
		cmdUpdateUserCollection())

	rootCmd.Execute()
}
