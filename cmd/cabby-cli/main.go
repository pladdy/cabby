package main

import (
	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/backends/sqlite"
	log "github.com/sirupsen/logrus"
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
	migrateVersion         int
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

func cmdMigrate() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate [up]",
		Short: "Migrate database up",
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

func dataStoreFromConfig(path string) cabby.DataStore {
	config := cabby.Config{}.Parse(path)
	ds, err := sqlite.NewDataStore(config.DataStore["path"])
	if err != nil {
		log.WithFields(log.Fields{"error": err, "path": config.DataStore["path"]}).Panic("Unable to connect to data store")
	}
	return ds
}

func main() {
	log.SetLevel(log.InfoLevel)

	// set up root and subcommands
	var rootCmd = &cobra.Command{Use: "cabby-cli"}
	rootCmd.PersistentFlags().StringVar(&configPath, "config", cabby.DefaultProductionConfig, "path to cabby config file")
	/* #nosec G104 */
	rootCmd.MarkFlagRequired("config")

	cmdCreate := cmdCreate()
	cmdDelete := cmdDelete()
	cmdMigrate := cmdMigrate()
	cmdUpdate := cmdUpdate()
	rootCmd.AddCommand(cmdCreate, cmdDelete, cmdMigrate, cmdUpdate)

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

	cmdMigrate.AddCommand(
		cmdMigrateUp())

	cmdUpdate.AddCommand(
		cmdUpdateAPIRoot(),
		cmdUpdateCollection(),
		cmdUpdateDiscovery(),
		cmdUpdateUser(),
		cmdUpdateUserCollection())

	err := rootCmd.Execute()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("Unable to execute command")
	}
}
