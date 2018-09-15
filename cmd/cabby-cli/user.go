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
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
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
			if userName == "" {
				log.Fatal("User name required")
			}
			if userPassword == "" {
				log.Fatal("Password required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&userName, "user", "u", "", "users name")
	cmd.MarkFlagRequired("user")
	cmd.PersistentFlags().StringVarP(&userPassword, "password", "p", "", "users password")
	cmd.MarkFlagRequired("password")
	cmd.PersistentFlags().BoolVarP(&userAdmin, "admin", "a", false, "user is an admin")

	return cmd
}

func cmdDeleteUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Delete a user",
		Long:  `delete user is used to delete a user from a server`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
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
			if userName == "" {
				log.Fatal("User name required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&userName, "user", "u", "", "users name")
	cmd.MarkFlagRequired("user")
	return cmd
}

func cmdUpdateUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Update a user",
		Long:  `update user is used to update a users properties`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
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
			if userName == "" {
				log.Fatal("User name required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&userName, "user", "u", "", "users name")
	cmd.MarkFlagRequired("user")
	cmd.PersistentFlags().BoolVarP(&userAdmin, "admin", "a", false, "user is an admin")
	cmd.MarkFlagRequired("admin")

	return cmd
}

func cmdCreateUserCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "userCollection",
		Short: "Create a user/collection assocation",
		Long:  `create UserCollection associates a user to a collection`,
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

			ca := cabby.CollectionAccess{ID: id, CanRead: userCollectionCanRead, CanWrite: userCollectionCanWrite}
			err = ds.UserService().CreateUserCollection(context.Background(), userName, ca)
			if err != nil {
				log.WithFields(log.Fields{"collection access": ca, "error": err, "user": userName}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if userName == "" {
				log.Fatal("User name required")
			}
			if collectionID == "" {
				log.Fatal("ID required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&userName, "user", "u", "", "users name")
	cmd.MarkFlagRequired("user")
	cmd.PersistentFlags().StringVarP(&collectionID, "id", "i", "", "users password")
	cmd.MarkFlagRequired("id")
	cmd.PersistentFlags().BoolVarP(&userCollectionCanRead, "can read", "r", false, "user can read from the collection")
	cmd.PersistentFlags().BoolVarP(&userCollectionCanWrite, "can write", "w", true, "user can write to the collection")

	return cmd
}

func cmdDeleteUserCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "userCollection",
		Short: "Delete a user/collection association",
		Long:  `delete a collection from a users collection access list`,
		Run: func(cmd *cobra.Command, args []string) {
			ds, err := dataStoreFromConfig(configPath, cabbyEnv)
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
			if userName == "" {
				log.Fatal("User name required")
			}
			if collectionID == "" {
				log.Fatal("ID required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&userName, "user", "u", "", "users name")
	cmd.MarkFlagRequired("user")
	cmd.PersistentFlags().StringVarP(&collectionID, "id", "i", "", "users password")
	cmd.MarkFlagRequired("id")
	return cmd
}

func cmdUpdateUserCollection() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "userCollection",
		Short: "Update a user/collection assocation",
		Long:  `update a collection access for a user`,
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

			ca := cabby.CollectionAccess{ID: id, CanRead: userCollectionCanRead, CanWrite: userCollectionCanWrite}
			err = ds.UserService().UpdateUserCollection(context.Background(), userName, ca)
			if err != nil {
				log.WithFields(log.Fields{"collection access": ca, "error": err, "user": userName}).Error("Failed to create")
			}
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if userName == "" {
				log.Fatal("User name required")
			}
			if collectionID == "" {
				log.Fatal("ID required")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&userName, "user", "u", "", "users name")
	cmd.MarkFlagRequired("user")
	cmd.PersistentFlags().StringVarP(&collectionID, "id", "i", "", "users password")
	cmd.MarkFlagRequired("id")
	cmd.PersistentFlags().BoolVarP(&userCollectionCanRead, "can read", "r", false, "user can read from the collection")
	cmd.PersistentFlags().BoolVarP(&userCollectionCanWrite, "can write", "w", true, "user can write to the collection")

	return cmd
}
