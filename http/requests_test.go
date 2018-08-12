package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetToken(t *testing.T) {
	tests := []struct {
		url      string
		index    int
		expected string
	}{
		{"/api_root/collections/collection_id/objects/stix_id", 0, ""},
		{"/api_root/collections/collection_id/objects/stix_id", 1, "api_root"},
		{"/api_root/collections/collection_id/objects/stix_id", 3, "collection_id"},
		{"/api_root/collections/collection_id/objects/stix_id", 5, "stix_id"},
		{"/api_root/collections/collection_id/objects/stix_id", 7, ""},
	}

	for _, test := range tests {
		result := getToken(test.url, test.index)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestLastURLPathToken(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/collections/", "collections"},
		{"/collections/someId", "someId"},
	}

	for _, test := range tests {
		result := lastURLPathToken(test.path)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

// func TestNewTaxiiRange(t *testing.T) {
// 	invalidRange := taxiiRange{first: -1, last: -1}
//
// 	tests := []struct {
// 		input       string
// 		resultRange taxiiRange
// 		isError     bool
// 	}{
// 		{"items 0-10", taxiiRange{first: 0, last: 10}, false},
// 		{"items 0 10", invalidRange, true},
// 		{"items 10", invalidRange, true},
// 		{"", invalidRange, false},
// 	}
//
// 	for _, test := range tests {
// 		result, err := newTaxiiRange(test.input)
// 		if result != test.resultRange {
// 			t.Error("Got:", result, "Expected:", test.resultRange)
// 		}
//
// 		if err != nil && test.isError == false {
// 			t.Error("Got:", err, "Expected: no error")
// 		}
// 	}
// }
//
// func TestTaxiiRangeValid(t *testing.T) {
// 	tests := []struct {
// 		tr       taxiiRange
// 		expected bool
// 	}{
// 		{taxiiRange{first: 1, last: 0}, false},
// 		{taxiiRange{first: 0, last: 0}, true},
// 	}
//
// 	for _, test := range tests {
// 		result := test.tr.Valid()
// 		if result != test.expected {
// 			t.Error("Got:", result, "Expected:", test.expected)
// 		}
// 	}
// }
//
// func TestTaxiiRangeString(t *testing.T) {
// 	tests := []struct {
// 		tr       taxiiRange
// 		expected string
// 	}{
// 		{taxiiRange{first: 0, last: 0}, "items 0-0"},
// 		{taxiiRange{first: 0, last: 0, total: 50}, "items 0-0/50"},
// 	}
//
// 	for _, test := range tests {
// 		result := test.tr.String()
// 		if result != test.expected {
// 			t.Error("Got:", result, "Expected:", test.expected)
// 		}
// 	}
// }

// func TestTakeCollectionAccessInvalidCollection(t *testing.T) {
// 	// create a request with a valid context BUT a path with an invalid collection in it
// 	request := withAuthContext(httptest.NewRequest("GET", "/foo/bar/baz", nil))
//
// 	tca := takeCollectionAccess(request)
// 	empty := taxiiCollectionAccess{}
//
// 	if tca != empty {
// 		t.Error("Got:", tca, "Expected:", empty)
// 	}
// }
//
// func TestTakeRequestRange(t *testing.T) {
// 	req := httptest.NewRequest("GET", "/test", nil)
//
// 	result := takeRequestRange(req)
// 	expected := taxiiRange{}
//
// 	if result != expected {
// 		t.Error("Got:", result, "Expected:", expected)
// 	}
// }
//
// func TestValidateUser(t *testing.T) {
// 	setupSQLite()
//
// 	ts := getStorer()
// 	defer ts.disconnect()
//
// 	tests := []struct {
// 		user     string
// 		pass     string
// 		expected bool
// 	}{
// 		{testUser, testPass, true},
// 		{"simon", "says", false},
// 	}
//
// 	for _, test := range tests {
// 		_, actual := validateUser(ts, test.user, test.pass)
// 		if actual != test.expected {
// 			t.Error("Got:", actual, "Expected:", test.expected)
// 		}
// 	}
// }
//
// func TestValidateUserFail(t *testing.T) {
// 	setupSQLite()
//
// 	s := getSQLiteDB()
// 	defer s.disconnect()
//
// 	_, err := s.db.Exec("drop table taxii_user")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ts := getStorer()
// 	defer ts.disconnect()
//
// 	_, isValid := validateUser(ts, "fail", "fail")
// 	if isValid {
// 		t.Error("Expected validation to be false")
// 	}
// }

func TestUserExists(t *testing.T) {
	tests := []struct {
		request  *http.Request
		expected bool
	}{
		{withAuthentication(httptest.NewRequest("GET", "/", nil)), true},
		{httptest.NewRequest("GET", "/", nil), false},
	}

	for _, test := range tests {
		result := userExists(test.request)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}
