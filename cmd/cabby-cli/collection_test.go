package main

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestCreateCollection(t *testing.T) {
	setUp()
	defer tearDown()

	command, resource := "create", "collection"
	expected := tester.Collection
	createTestUser(tester.Collection.ID.String())

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{
			command, resource, "--config", CLIConfig,
			"-a", expected.APIRootPath,
			"-i", expected.ID.String(),
			"-t", expected.Title,
			"-d", expected.Description}, false},
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
			ctx := cabby.WithUser(context.Background(), tester.User)
			result, _ := ds.CollectionService().Collection(ctx, expected.APIRootPath, expected.ID.String())

			passed := tester.CompareCollection(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}

func TestDeleteCollection(t *testing.T) {
	setUp()
	defer tearDown()

	command, resource := "delete", "collection"
	toDelete := tester.Collection
	expected := cabby.Collection{}

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{command, resource, "--config", CLIConfig, "-i", toDelete.ID.String()}, false},
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
			result, _ := ds.CollectionService().Collection(context.Background(), toDelete.APIRootPath, toDelete.ID.String())

			passed := tester.CompareCollection(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}

func TestUpdateCollection(t *testing.T) {
	setUp()
	defer tearDown()

	createTestUser(tester.Collection.ID.String())
	expected := tester.Collection

	// create a collection to modify
	cmd := exec.Command(CLICommand, "create", "collection", "--config", CLIConfig,
		"-a", expected.APIRootPath,
		"-i", expected.ID.String(),
		"-t", expected.Title,
		"-d", expected.Description)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// test updates
	expected.APIRootPath = "/updated/api/root/path/"
	expected.Title = "an updated title"
	expected.Description = "an updated description"

	command, resource := "update", "collection"

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{command, resource, "--config", CLIConfig,
			"-a", expected.APIRootPath,
			"-i", expected.ID.String(),
			"-t", expected.Title,
			"-d", expected.Description}, false},
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
			ctx := cabby.WithUser(context.Background(), tester.User)
			result, _ := ds.CollectionService().Collection(ctx, expected.APIRootPath, expected.ID.String())

			passed := tester.CompareCollection(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}
