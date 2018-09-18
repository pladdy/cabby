package main

import (
	"context"
	"os"
	"os/exec"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestCreateDiscovery(t *testing.T) {
	setUp()
	defer tearDown()

	command, resource := "create", "discovery"
	expected := tester.Discovery

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{
			command, resource,
			"--config", CLIConfig,
			"-t", expected.Title,
			"-d", expected.Description,
			"-c", expected.Contact,
			"-u", expected.Default}, false},
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
			result, _ := ds.DiscoveryService().Discovery(context.Background())

			passed := tester.CompareDiscovery(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}

func TestDeleteDiscovery(t *testing.T) {
	setUp()
	defer tearDown()

	cmd := exec.Command(CLICommand, "delete", "discovery", "--config", CLIConfig)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		t.Error(err)
	}

	expected := cabby.Discovery{}
	ds := testDataStore()
	result, _ := ds.DiscoveryService().Discovery(context.Background())

	passed := tester.CompareDiscovery(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}

}

func TestUpdateDiscovery(t *testing.T) {
	setUp()
	defer tearDown()

	expected := tester.Discovery

	// create a discovery to modify
	cmd := exec.Command(CLICommand, "create", "discovery",
		"--config", CLIConfig,
		"-t", expected.Title,
		"-d", expected.Description,
		"-c", expected.Contact,
		"-u", expected.Default)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// test updates
	expected.Title = "an updated title"
	expected.Description = "an updated description"
	expected.Contact = "a new contact"
	expected.Default = "https://newurl:8042/taxii"

	command, resource := "update", "discovery"

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{command, resource,
			"--config", CLIConfig,
			"-t", expected.Title,
			"-d", expected.Description,
			"-c", expected.Contact,
			"-u", expected.Default}, false},
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
			result, _ := ds.DiscoveryService().Discovery(context.Background())

			passed := tester.CompareDiscovery(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}
