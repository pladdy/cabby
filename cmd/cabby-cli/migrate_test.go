package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestMigrateUp(t *testing.T) {
	setUp()
	defer tearDown()

	command, direction := "migrate", "up"

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, direction, "--config", CLIConfig}, false},
	}

	for _, test := range tests {
		cmd := exec.Command(CLICommand, test.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout

		err := cmd.Run()
		if test.expectError && err == nil {
			t.Error("Expected error: no parameters set")
		}

		if !test.expectError {
			ds := testDataStore()
			result, _ := ds.MigrationService().CurrentVersion()

			if result != 1 {
				t.Error("Expected schema verstion to be 1")
			}
		}
	}
}
