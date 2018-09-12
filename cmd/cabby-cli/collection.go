package main

import (
	"context"

	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func cmdCreateCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collection",
		Short: "Create a collection",
		Long:  `create collection is used to create a collection on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			id, err := cabby.IDFromString(collectionID)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "id": collectionID}).Error("Failed to create ID")
			}
			newCollection := cabby.Collection{ID: id, Title: collectionTitle, Description: collectionDescription}

			err = ds.CollectionService().CreateCollection(context.Background(), newCollection)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "collection": newCollection}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if collectionID == "" {
				log.Fatal("ID required")
			}
			if collectionTitle == "" {
				log.Fatal("Title required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&collectionID, "id", "i", "", "collection id")
	cmd.MarkFlagRequired("id")
	cmd.PersistentFlags().StringVarP(&collectionTitle, "title", "t", "", "collection title")
	cmd.MarkFlagRequired("title")
	cmd.PersistentFlags().StringVarP(&collectionDescription, "description", "d", "", "collection description")

	return cmd
}

func cmdDeleteCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collection",
		Short: "Delete a collection",
		Long:  `delete collection is used to delete a collection from a server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			err = ds.CollectionService().DeleteCollection(context.Background(), collectionID)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "id": collectionID}).Error("Failed to delete")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if collectionID == "" {
				log.Fatal("ID required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&collectionID, "id", "i", "", "collection id")
	cmd.MarkFlagRequired("id")

	return cmd
}

func cmdUpdateCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collection",
		Short: "Update a collection",
		Long:  `update collection is used to update a collection on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			id, err := cabby.IDFromString(collectionID)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "id": id}).Error("Failed to create ID")
			}
			newCollection := cabby.Collection{ID: id, Title: collectionTitle, Description: collectionDescription}

			err = ds.CollectionService().UpdateCollection(context.Background(), newCollection)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "collection": newCollection}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if collectionID == "" {
				log.Fatal("ID required")
			}
			if collectionTitle == "" {
				log.Fatal("Title required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&collectionID, "id", "i", "", "collection id")
	cmd.MarkFlagRequired("id")
	cmd.PersistentFlags().StringVarP(&collectionTitle, "title", "t", "", "collection title")
	cmd.MarkFlagRequired("title")
	cmd.PersistentFlags().StringVarP(&collectionDescription, "description", "d", "", "collection description")

	return cmd
}
