package main

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestCreateUser(t *testing.T) {
	setUp()
	defer tearDown()

	command, resource := "create", "user"
	expected := tester.User

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{
			command, resource, "--config", CLIConfig,
			"-u", tester.UserEmail,
			"-p", tester.UserPassword,
			"-a", strconv.FormatBool(expected.CanAdmin)}, false},
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
			result, _ := ds.UserService().User(context.Background(), tester.UserEmail, tester.UserPassword)

			passed := tester.CompareUser(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}

func TestDeleteUser(t *testing.T) {
	setUp()
	defer tearDown()

	command, resource := "delete", "user"
	expected := cabby.User{}

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{command, resource, "--config", CLIConfig, "-u", tester.UserEmail}, false},
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
			result, _ := ds.UserService().User(context.Background(), tester.UserEmail, tester.UserPassword)

			passed := tester.CompareUser(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}

func TestUpdateUser(t *testing.T) {
	setUp()
	defer tearDown()

	expected := tester.User

	// create a user to modify
	cmd := exec.Command(CLICommand, "create", "user", "--config", CLIConfig,
		"-u", tester.UserEmail,
		"-p", tester.UserPassword,
		"-a", strconv.FormatBool(expected.CanAdmin))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// test updates
	// Note: to disable admin you don't set the flag (defaults to false)
	expected.CanAdmin = false

	command, resource := "update", "user"

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{
			command, resource, "--config", CLIConfig,
			"-u", tester.UserEmail}, false},
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
			result, _ := ds.UserService().User(context.Background(), tester.UserEmail, tester.UserPassword)

			passed := tester.CompareUser(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}

func TestCreateUserCollection(t *testing.T) {
	setUp()
	defer tearDown()

	createTestUser(tester.Collection.ID.String())
	command, resource := "create", "userCollection"
	expected := tester.UserCollectionList

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{
			command, resource, "--config", CLIConfig,
			"-u", tester.UserEmail,
			"-i", tester.CollectionID,
			"-r",
			"-w"}, false},
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
			result, _ := ds.UserService().UserCollections(ctx, tester.UserEmail)

			rca, ok := result.CollectionAccessList[tester.Collection.ID]
			if !ok {
				t.Error("Got:", rca, "Expected:", expected)
			}

			eca := expected.CollectionAccessList[tester.Collection.ID]
			if rca.CanRead != eca.CanRead {
				t.Error("Got:", rca.CanRead, "Expected:", eca.CanRead)
			}
			if rca.CanWrite != eca.CanWrite {
				t.Error("Got:", rca.CanWrite, "Expected:", eca.CanWrite)
			}
		}
	}
}

func TestDeleteUserCollection(t *testing.T) {
	setUp()
	defer tearDown()

	command, resource := "delete", "userCollection"

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{command, resource, "--config", CLIConfig, "-u", tester.UserEmail}, false},
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
			result, _ := ds.UserService().UserCollections(context.Background(), tester.UserEmail)

			_, ok := result.CollectionAccessList[tester.Collection.ID]
			if ok {
				t.Error("Expected list to be empty")
			}
		}
	}
}

func TestUpdateUserCollection(t *testing.T) {
	setUp()
	defer tearDown()

	expected := tester.UserCollectionList
	createTestUser(tester.Collection.ID.String())

	// create a userCollection to modify
	cmd := exec.Command(CLICommand, "create", "userCollection", "--config", CLIConfig,
		"-u", tester.UserEmail,
		"-i", tester.CollectionID,
		"-r", "true",
		"-w", "true")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// test updates
	ca := expected.CollectionAccessList[tester.Collection.ID]
	ca.CanRead = false
	ca.CanWrite = false
	expected.CollectionAccessList[tester.Collection.ID] = ca

	command, resource := "update", "userCollection"

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{
			command, resource, "--config", CLIConfig,
			"-u", tester.UserEmail,
			"-i", tester.CollectionID}, false},
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
			result, _ := ds.UserService().UserCollections(ctx, tester.UserEmail)

			rca, ok := result.CollectionAccessList[tester.Collection.ID]
			if !ok {
				t.Error("Got:", rca, "Expected:", expected)
			}

			eca := expected.CollectionAccessList[tester.Collection.ID]
			if rca.CanRead != eca.CanRead {
				t.Error("Got:", rca.CanRead, "Expected:", eca.CanRead)
			}
			if rca.CanWrite != eca.CanWrite {
				t.Error("Got:", rca.CanWrite, "Expected:", eca.CanWrite)
			}
		}
	}
}
