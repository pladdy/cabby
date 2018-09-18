package main

import (
	"context"

	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func cmdCreateDiscovery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "Create the discovery resource",
		Long:  `create discovery is used to create the discovery resource on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			newDiscovery := cabby.Discovery{
				Title:       discoveryTitle,
				Description: discoveryDescription,
				Contact:     discoveryContact,
				Default:     discoveryDefault}

			err = ds.DiscoveryService().CreateDiscovery(context.Background(), newDiscovery)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "api_root": newDiscovery}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if discoveryTitle == "" {
				log.Fatal("Title required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&discoveryTitle, "title", "t", "", "title of the server")
	cmd.MarkFlagRequired("title")
	cmd.PersistentFlags().StringVarP(&discoveryDescription, "description", "d", "", "discovery description")
	cmd.PersistentFlags().StringVarP(&discoveryContact, "contact", "c", "", "contact for server")
	cmd.PersistentFlags().StringVarP(&discoveryDefault, "default", "u", "", "default URL for server")

	return cmd
}

func cmdDeleteDiscovery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "Delete the discovery resource",
		Long:  `delete discovery is used to delete the discovery from a server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			err = ds.DiscoveryService().DeleteDiscovery(context.Background())
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to delete")
			}
		},
	}

	return cmd
}

func cmdUpdateDiscovery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "Update the discovery resource",
		Long:  `update discovery is used to update the discovery resource on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			newDiscovery := cabby.Discovery{
				Title:       discoveryTitle,
				Description: discoveryDescription,
				Contact:     discoveryContact,
				Default:     discoveryDefault}

			err = ds.DiscoveryService().UpdateDiscovery(context.Background(), newDiscovery)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "discovery": newDiscovery}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if discoveryTitle == "" {
				log.Fatal("Title required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&discoveryTitle, "title", "t", "", "title of the server")
	cmd.MarkFlagRequired("title")
	cmd.PersistentFlags().StringVarP(&discoveryDescription, "description", "d", "", "discovery description")
	cmd.PersistentFlags().StringVarP(&discoveryContact, "contact", "c", "", "contact for server")
	cmd.PersistentFlags().StringVarP(&discoveryDefault, "default", "u", "", "default URL for server")

	return cmd
}
