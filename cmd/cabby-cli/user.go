package main

import (
	"context"

	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func cmdCreateUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user -u <user> -p <password>",
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
		Use:   "user -u <user>",
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
	cmd.PersistentFlags().BoolVarP(&userAdmin, "admin", "a", false, "user is an admin")

	return cmd
}
