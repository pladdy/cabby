package main

import (
	"net/http/httptest"
	"testing"
)

func TestTakeCollectionAccessInvalidCollection(t *testing.T) {
	// create a request with a valid context BUT a path with an invalid collection in it
	request := withAuthContext(httptest.NewRequest("GET", "/foo/bar/baz", nil))

	tca := takeCollectionAccess(request)
	empty := taxiiCollectionAccess{}

	if tca != empty {
		t.Error("Got:", tca, "Expected:", empty)
	}
}

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
