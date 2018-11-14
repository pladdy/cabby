package main

import (
	"context"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func cmdCreateCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collection",
		Short: "Create a collection",
		Long:  `create collection is used to create a collection on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds := dataStoreFromConfig(configPath)
			defer ds.Close()

			id, err := cabby.IDFromString(collectionID)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "id": collectionID}).Error("Failed to create ID")
			}
			newCollection := cabby.Collection{
				APIRootPath: apiRootPath,
				ID:          id,
				Title:       collectionTitle,
				Description: collectionDescription}

			err = ds.CollectionService().CreateCollection(context.Background(), newCollection)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "collection": newCollection}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateCollectionFlags()
		},
	}

	return withCollectionFlags(cmd)
}

func cmdDeleteCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collection",
		Short: "Delete a collection",
		Long:  `delete collection is used to delete a collection from a server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds := dataStoreFromConfig(configPath)
			defer ds.Close()

			err := ds.CollectionService().DeleteCollection(context.Background(), collectionID)
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

	return withCollectionIDFlag(cmd)
}

func cmdUpdateCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collection",
		Short: "Update a collection",
		Long:  `update collection is used to update a collection on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds := dataStoreFromConfig(configPath)
			defer ds.Close()

			id, err := cabby.IDFromString(collectionID)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "id": id}).Error("Failed to create ID")
			}
			newCollection := cabby.Collection{
				APIRootPath: apiRootPath,
				ID:          id,
				Title:       collectionTitle,
				Description: collectionDescription}

			err = ds.CollectionService().UpdateCollection(context.Background(), newCollection)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "collection": newCollection}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateCollectionFlags()
		},
	}

	return withCollectionFlags(cmd)
}

func validateCollectionFlags() {
	if apiRootPath == "" {
		log.Fatal("API Root Path required")
	}
	if collectionID == "" {
		log.Fatal("ID required")
	}
	if collectionTitle == "" {
		log.Fatal("Title required")
	}
}
