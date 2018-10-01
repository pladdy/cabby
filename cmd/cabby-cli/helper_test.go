package main

import (
	"os"
	"os/exec"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/sqlite"
	"github.com/pladdy/cabby/tester"
	log "github.com/sirupsen/logrus"
)

const (
	CLICommand = "./cabby-cli"
	CLIConfig  = "./cabby-cli-config.json"
)

func createTestUser(collectionID string) {
	cmd := exec.Command(CLICommand, "create", "user",
		"--config", CLIConfig, "-u", tester.User.Email, "-p", tester.UserPassword)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command(CLICommand, "create", "userCollection",
		"--config", CLIConfig, "-u", tester.User.Email, "-i", collectionID, "-r", "true", "-w", "true")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func testDataStore() cabby.DataStore {
	config := cabby.Config{}.Parse(CLIConfig)

	ds, err := sqlite.NewDataStore(config.DataStore["path"])
	if err != nil {
		log.Fatal(err)
	}
	return ds
}
