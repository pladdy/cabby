package main

import (
	"context"
	"strings"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func cmdCreateAPIRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiRoot",
		Short: "Create a apiRoot",
		Long:  `create apiRoot is used to create a apiRoot on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds := dataStoreFromConfig(configPath)
			defer ds.Close()

			newAPIRoot := cabby.APIRoot{
				Path:             apiRootPath,
				Title:            apiRootTitle,
				Description:      apiRootDescription,
				Versions:         strings.Split(apiRootVersions, ","),
				MaxContentLength: maxContentLength}

			err := ds.APIRootService().CreateAPIRoot(context.Background(), newAPIRoot)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "api_root": newAPIRoot}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateAPIRootFlags()
		},
	}

	return withAPIRootFlags(cmd)
}

func cmdDeleteAPIRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiRoot",
		Short: "Delete a apiRoot",
		Long:  `delete apiRoot is used to delete a apiRoot from a server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds := dataStoreFromConfig(configPath)
			defer ds.Close()

			err := ds.APIRootService().DeleteAPIRoot(context.Background(), apiRootPath)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to delete")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if apiRootPath == "" {
				log.Fatal("API Root Path required")
			}
		},
	}

	return withAPIRootPathFlag(cmd)
}

func cmdUpdateAPIRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiRoot",
		Short: "Update a apiRoot",
		Long:  `update apiRoot is used to update a apiRoot on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds := dataStoreFromConfig(configPath)
			defer ds.Close()

			newAPIRoot := cabby.APIRoot{
				Path:             apiRootPath,
				Title:            apiRootTitle,
				Description:      apiRootDescription,
				Versions:         strings.Split(apiRootVersions, ","),
				MaxContentLength: maxContentLength}

			err := ds.APIRootService().CreateAPIRoot(context.Background(), newAPIRoot)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "apiRoot": newAPIRoot}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateAPIRootFlags()
		},
	}

	return withAPIRootFlags(cmd)
}

func validateAPIRootFlags() {
	if apiRootPath == "" {
		log.Fatal("API Root Path required")
	}
	if apiRootTitle == "" {
		log.Fatal("Title required")
	}
	if apiRootVersions == "" {
		log.Fatal("Version(s) required")
	}
	if maxContentLength <= 0 {
		log.Fatal("Max Content Length required")
	}
}
