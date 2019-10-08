package main

import (
	"os"
	"os/exec"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/backends/sqlite"
	"github.com/pladdy/cabby/tester"
	log "github.com/sirupsen/logrus"
)

const (
	CLICommand = "./cabby-cli"
	CLIConfig  = "./cabby-cli-config.json"
)

var (
	commands    = []string{"create", "delete", "update"}
	subCommands = []string{"apiRoot", "collection", "discovery", "user", "userCollection"}
)

func init() {
	cmd := exec.Command("go", "build", "-tags", "json1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

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

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func setUp() {
	ds := testDataStore()
	defer ds.Close()

	commandsToRun := []struct {
		command string
		args    []string
	}{
		{"./cabby-cli", []string{"migrate", "up", "--config", CLIConfig}},
	}

	for _, command := range commandsToRun {
		cmd := exec.Command(command.command, command.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout

		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func tearDown() {
	filesToRemove := []string{"cabby-cli.db"}

	for _, file := range filesToRemove {
		err := os.Remove(file)
		if err != nil {
			log.Warn(err)
		}
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
