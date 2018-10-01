package main

import (
	"github.com/pladdy/cabby"
	"github.com/spf13/cobra"
)

const eightMB = 8388608

/* api root flags */

func withAPIRootFlags(cmd *cobra.Command) *cobra.Command {
	cmd = withAPIRootDescriptionFlag(cmd)
	cmd = withAPIRootMaxContentLengthFlag(cmd)
	cmd = withAPIRootPathFlag(cmd)
	cmd = withAPIRootTitleFlag(cmd)
	return withAPIRootVersionsFlag(cmd)
}

func withAPIRootDescriptionFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&apiRootDescription, "description", "d", "", "api root description")
	return cmd
}

func withAPIRootMaxContentLengthFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().Int64VarP(&maxContentLength, "max_content_length", "m", eightMB, "max content length of requests supported")
	cmd.MarkFlagRequired("max_content_length")
	return cmd
}

func withAPIRootPathFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&apiRootPath, "api_root_path", "a", "", "path for the api root")
	cmd.MarkFlagRequired("api_root_path")
	return cmd
}

func withAPIRootTitleFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&apiRootTitle, "title", "t", "", "title of api root")
	cmd.MarkFlagRequired("title")
	return cmd
}

func withAPIRootVersionsFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&apiRootVersions, "versions", "v", cabby.TaxiiVersion, "versions api root supports")
	cmd.MarkFlagRequired("versions")
	return cmd
}

/* collection flags */

func withCollectionFlags(cmd *cobra.Command) *cobra.Command {
	cmd = withAPIRootPathFlag(cmd)
	cmd = withCollectionIDFlag(cmd)
	cmd = withCollectionTitleFlag(cmd)
	return withCollectionDescriptionFlag(cmd)
}

func withCollectionDescriptionFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&collectionDescription, "description", "d", "", "collection description")
	return cmd
}

func withCollectionIDFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&collectionID, "id", "i", "", "collection id")
	cmd.MarkFlagRequired("id")
	return cmd
}

func withCollectionTitleFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&collectionTitle, "title", "t", "", "collection title")
	cmd.MarkFlagRequired("title")
	return cmd
}

/* discovery flags */

func withDiscoveryContactFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&discoveryContact, "contact", "c", "", "contact for server")
	return cmd
}

func withDiscoveryDefaultsFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&discoveryDefault, "default", "u", "", "default URL for server")
	return cmd
}

func withDiscoveryDescriptionFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&discoveryDescription, "description", "d", "", "discovery description")
	return cmd
}

func withDiscoveryTitleFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&discoveryTitle, "title", "t", "", "title of the server")
	cmd.MarkFlagRequired("title")
	return cmd
}

/* user flags */

func withAdminFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().BoolVarP(&userAdmin, "admin", "a", false, "user is an admin")
	return cmd
}

func withPasswordFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&userPassword, "password", "p", "", "user's password")
	cmd.MarkFlagRequired("password")
	return cmd
}

func withReadWriteFlags(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().BoolVarP(&userCollectionCanRead, "read", "r", false, "user can read from the collection")
	cmd.PersistentFlags().BoolVarP(&userCollectionCanWrite, "write", "w", false, "user can write to the collection")
	return cmd
}

func withUserFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringVarP(&userName, "user", "u", "", "user's name")
	cmd.MarkFlagRequired("user")
	return cmd
}
