package main

import (
	"os"
	"os/exec"
	"testing"

	log "github.com/sirupsen/logrus"
)

const (
	RelativeSchemaPath = "../../sqlite/schema.sql"
)

var (
	commands    = []string{"create", "delete", "update"}
	subCommands = []string{"apiRoot", "collection", "discovery", "user", "userCollection"}
)

func init() {
	cmd := exec.Command("go", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
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
	commandsToRun := []struct {
		command      string
		args         []string
		expectedFile string
	}{
		{"sqlite3", []string{"cabby-cli.db", ".read " + RelativeSchemaPath}, "cabby-cli.db"},
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

func TestCommands(t *testing.T) {
	setUp()
	defer tearDown()

	for _, command := range commands {
		cmd := exec.Command("./cabby-cli", command)
		cmd.Stderr = os.Stdout

		err := cmd.Run()
		if err != nil {
			t.Error("Got:", err, "Expected: nil", "For:", command)
		}
	}
}

func TestSubCommands(t *testing.T) {
	setUp()
	defer tearDown()

	for _, command := range commands {
		for _, subCommand := range subCommands {
			cmd := exec.Command(CLICommand, command, subCommand, "-h")
			cmd.Stderr = os.Stdout

			err := cmd.Run()
			if err != nil {
				t.Error("Got:", err, "Expected: nil", "Command:", command, "SubCommand:", subCommand)
			}
		}
	}
}
