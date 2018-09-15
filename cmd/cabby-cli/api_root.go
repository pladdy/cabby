package main

import (
	"context"
	"strings"

	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const eightMB = 8388608

func cmdCreateAPIRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiRoot",
		Short: "Create a apiRoot",
		Long:  `create apiRoot is used to create a apiRoot on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			newAPIRoot := cabby.APIRoot{
				Path:             apiRootPath,
				Title:            apiRootTitle,
				Description:      apiRootDescription,
				Versions:         strings.Split(apiRootVersions, ","),
				MaxContentLength: maxContentLength}

			err = ds.APIRootService().CreateAPIRoot(context.Background(), newAPIRoot)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "api_root": newAPIRoot}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
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
		},
	}

	cmd.PersistentFlags().StringVarP(&apiRootPath, "api_root_path", "a", "", "path for the api root")
	cmd.MarkFlagRequired("api_root_path")
	cmd.PersistentFlags().StringVarP(&apiRootTitle, "title", "t", "", "title of api root")
	cmd.MarkFlagRequired("title")
	cmd.PersistentFlags().StringVarP(&apiRootVersions, "versions", "v", cabby.TaxiiVersion, "versions api root supports")
	cmd.MarkFlagRequired("versions")
	cmd.PersistentFlags().Int64VarP(&maxContentLength, "max_content_length", "m", eightMB, "max content length of requests supported")
	cmd.MarkFlagRequired("max_content_length")
	cmd.PersistentFlags().StringVarP(&apiRootDescription, "description", "d", "", "api root description")

	return cmd
}

func cmdDeleteAPIRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiRoot",
		Short: "Delete a apiRoot",
		Long:  `delete apiRoot is used to delete a apiRoot from a server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			err = ds.APIRootService().DeleteAPIRoot(context.Background(), apiRootPath)
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

	cmd.PersistentFlags().StringVarP(&apiRootPath, "api_root_path", "a", "", "path for the api root")
	cmd.MarkFlagRequired("api_root_path")

	return cmd
}

func cmdUpdateAPIRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiRoot",
		Short: "Update a apiRoot",
		Long:  `update apiRoot is used to update a apiRoot on the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			newAPIRoot := cabby.APIRoot{
				Path:             apiRootPath,
				Title:            apiRootTitle,
				Description:      apiRootDescription,
				Versions:         strings.Split(apiRootVersions, ","),
				MaxContentLength: maxContentLength}

			err = ds.APIRootService().CreateAPIRoot(context.Background(), newAPIRoot)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "apiRoot": newAPIRoot}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
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
		},
	}

	cmd.PersistentFlags().StringVarP(&apiRootPath, "api_root_path", "a", "", "path for the api root")
	cmd.MarkFlagRequired("api_root_path")
	cmd.PersistentFlags().StringVarP(&apiRootTitle, "title", "t", "", "title of api root")
	cmd.MarkFlagRequired("title")
	cmd.PersistentFlags().StringVarP(&apiRootVersions, "versions", "v", cabby.TaxiiVersion, "versions api root supports")
	cmd.MarkFlagRequired("versions")
	cmd.PersistentFlags().Int64VarP(&maxContentLength, "max_content_length", "m", eightMB, "max content length of requests supported")
	cmd.MarkFlagRequired("max_content_length")
	cmd.PersistentFlags().StringVarP(&apiRootDescription, "description", "d", "", "api root description")

	return cmd
}
