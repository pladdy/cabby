package main

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/pladdy/cabby/tester"
)

func TestCreateAPIRoot(t *testing.T) {
	setUp()
	defer tearDown()

	command, resource := "create", "apiRoot"
	expected := tester.APIRoot

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{
			command, resource,
			"--config", CLIConfig,
			"-a", expected.Path,
			"-t", expected.Title,
			"-d", expected.Description,
			"-v", strings.Join(expected.Versions, ","),
			"-m", strconv.Itoa(int(expected.MaxContentLength))}, false},
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
			result, _ := ds.APIRootService().APIRoot(context.Background(), expected.Path)

			passed := tester.CompareAPIRoot(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}

func TestDeleteAPIRoot(t *testing.T) {
	setUp()
	defer tearDown()

	command, resource := "delete", "apiRoot"
	expected := tester.APIRoot

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{command, resource, "--config", CLIConfig, "-a", expected.Path}, false},
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
			result, _ := ds.APIRootService().APIRoot(context.Background(), expected.Path)

			passed := tester.CompareAPIRoot(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}

func TestUpdateAPIRoot(t *testing.T) {
	setUp()
	defer tearDown()

	expected := tester.APIRoot

	// create a apiRoot to modify
	cmd := exec.Command(CLICommand, "create", "apiRoot",
		"--config", CLIConfig,
		"-a", expected.Path,
		"-t", expected.Title,
		"-d", expected.Description,
		"-v", strings.Join(expected.Versions, ","),
		"-m", strconv.Itoa(int(expected.MaxContentLength)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// test updates
	expected.Path = "/updated/api/root/path/"
	expected.Title = "an updated title"
	expected.Description = "an updated description"
	expected.Versions = []string{"taxii-2.0", "taxii-2.1"}
	expected.MaxContentLength = 8192

	command, resource := "update", "apiRoot"

	tests := []struct {
		args        []string
		expectError bool
	}{
		{[]string{command, resource}, true},
		{[]string{command, resource, "--config", CLIConfig}, true},
		{[]string{command, resource,
			"--config", CLIConfig,
			"-a", expected.Path,
			"-t", expected.Title,
			"-d", expected.Description,
			"-v", strings.Join(expected.Versions, ","),
			"-m", strconv.Itoa(int(expected.MaxContentLength))}, false},
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
			result, _ := ds.APIRootService().APIRoot(context.Background(), expected.Path)

			passed := tester.CompareAPIRoot(result, expected)
			if !passed {
				t.Error("Comparison failed")
			}
		}
	}
}
