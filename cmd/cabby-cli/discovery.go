package main

import (
	"context"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func cmdCreateDiscovery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "Create the discovery resource",
		Long:  `create discovery is used to create the discovery resource on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds := dataStoreFromConfig(configPath)
			defer ds.Close()

			newDiscovery := cabby.Discovery{
				Title:       discoveryTitle,
				Description: discoveryDescription,
				Contact:     discoveryContact,
				Default:     discoveryDefault}

			err := ds.DiscoveryService().CreateDiscovery(context.Background(), newDiscovery)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "discovery": newDiscovery}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateDiscoveryFlags()
		},
	}

	cmd = withDiscoveryContactFlag(cmd)
	cmd = withDiscoveryDefaultsFlag(cmd)
	cmd = withDiscoveryDescriptionFlag(cmd)
	return withDiscoveryTitleFlag(cmd)
}

func cmdDeleteDiscovery() *cobra.Command {
	return &cobra.Command{
		Use:   "discovery",
		Short: "Delete the discovery resource",
		Long:  `delete discovery is used to delete the discovery from a server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds := dataStoreFromConfig(configPath)
			defer ds.Close()

			err := ds.DiscoveryService().DeleteDiscovery(context.Background())
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to delete")
			}
		},
	}
}

func cmdUpdateDiscovery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "Update the discovery resource",
		Long:  `update discovery is used to update the discovery resource on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds := dataStoreFromConfig(configPath)
			defer ds.Close()

			newDiscovery := cabby.Discovery{
				Title:       discoveryTitle,
				Description: discoveryDescription,
				Contact:     discoveryContact,
				Default:     discoveryDefault}

			err := ds.DiscoveryService().UpdateDiscovery(context.Background(), newDiscovery)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "discovery": newDiscovery}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateDiscoveryFlags()
		},
	}

	cmd = withDiscoveryContactFlag(cmd)
	cmd = withDiscoveryDefaultsFlag(cmd)
	cmd = withDiscoveryDescriptionFlag(cmd)
	return withDiscoveryTitleFlag(cmd)
}

func validateDiscoveryFlags() {
	if discoveryTitle == "" {
		log.Fatal("Title required")
	}
}
