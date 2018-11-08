package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func cmdMigrateUp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Migrate the database up",
		Long:  `migrate up migrates the database to the most recent version`,
		PreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.InfoLevel)
		},
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			err = ds.MigrationService().Up()
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to create")
			}
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.ErrorLevel)
		},
	}

	return cmd
}
