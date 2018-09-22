package main

import (
	"context"

	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func cmdCreateUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Create a user",
		Long:  `create user is used to create a user that will access the server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			newUser := cabby.User{Email: userName, CanAdmin: userAdmin}
			err = ds.UserService().CreateUser(context.Background(), newUser, userPassword)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "user": newUser}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateUserFlags()
			if userPassword == "" {
				log.Fatal("Password required")
			}
		},
	}

	cmd = withUserFlag(cmd)
	cmd = withPasswordFlag(cmd)
	return withAdminFlag(cmd)
}

func cmdDeleteUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Delete a user",
		Long:  `delete user is used to delete a user from a server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			err = ds.UserService().DeleteUser(context.Background(), userName)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "user": userName}).Error("Failed to delete")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateUserFlags()
		},
	}

	return withUserFlag(cmd)
}

func cmdUpdateUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Update a user",
		Long:  `update user is used to update a users properties`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			newUser := cabby.User{Email: userName, CanAdmin: userAdmin}
			err = ds.UserService().UpdateUser(context.Background(), newUser)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "user": newUser}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateUserFlags()
		},
	}

	cmd = withUserFlag(cmd)
	cmd = withAdminFlag(cmd)
	cmd.MarkFlagRequired("admin")
	return cmd
}

func cmdCreateUserCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "userCollection",
		Short: "Create a user/collection assocation",
		Long:  `create UserCollection associates a user to a collection`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			id, err := cabby.IDFromString(collectionID)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "id": collectionID}).Error("Failed to create ID")
			}

			ca := cabby.CollectionAccess{
				ID:       id,
				CanRead:  userCollectionCanRead,
				CanWrite: userCollectionCanWrite}

			err = ds.UserService().CreateUserCollection(context.Background(), userName, ca)
			if err != nil {
				log.WithFields(log.Fields{"collection access": ca, "error": err, "user": userName}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateUserCollectionFlags()
			validateUserCollectionFlags()
		},
	}

	cmd = withUserFlag(cmd)
	cmd = withCollectionIDFlag(cmd)
	return withReadWriteFlags(cmd)
}

func cmdDeleteUserCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "userCollection",
		Short: "Delete a user/collection association",
		Long:  `delete a collection from a users collection access list`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			err = ds.UserService().DeleteUserCollection(context.Background(), userName, collectionID)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "user": userName}).Error("Failed to delete")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateUserCollectionFlags()
			validateUserCollectionFlags()
		},
	}

	cmd = withUserFlag(cmd)
	cmd = withCollectionIDFlag(cmd)
	return cmd
}

func cmdUpdateUserCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "userCollection",
		Short: "Update a user/collection assocation",
		Long:  `update a collection access for a user`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
			}
			defer ds.Close()

			id, err := cabby.IDFromString(collectionID)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "id": collectionID}).Error("Failed to create ID")
			}

			ca := cabby.CollectionAccess{
				ID:       id,
				CanRead:  userCollectionCanRead,
				CanWrite: userCollectionCanWrite}

			err = ds.UserService().UpdateUserCollection(context.Background(), userName, ca)
			if err != nil {
				log.WithFields(log.Fields{"collection access": ca, "error": err, "user": userName}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateUserCollectionFlags()
			validateUserCollectionFlags()
		},
	}

	cmd = withUserFlag(cmd)
	cmd = withCollectionIDFlag(cmd)
	return withReadWriteFlags(cmd)
}

func validateUserFlags() {
	if userName == "" {
		log.Fatal("User name required")
	}
}

func validateUserCollectionFlags() {
	if collectionID == "" {
		log.Fatal("ID required")
	}
}
