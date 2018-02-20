package main

import "testing"

func TestValidateUser(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	tests := []struct {
		user     string
		pass     string
		expected bool
	}{
		{testUser, testPass, true},
		{"simon", "says", false},
	}

	for _, test := range tests {
		_, actual := validateUser(ts, test.user, test.pass)
		if actual != test.expected {
			t.Error("Got:", actual, "Expected:", test.expected)
		}
	}
}
