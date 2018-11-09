package main

import (
	"os"
	"os/exec"
	"testing"
)

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
