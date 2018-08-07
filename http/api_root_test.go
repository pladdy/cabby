package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cabby "github.com/pladdy/cabby2"
)

func TestAPIRootHandlerGet(t *testing.T) {
	ds := APIRootService{}
	ds.APIRootFn = func(path string) (cabby.APIRoot, error) {
		return testAPIRoot(), nil
	}

	// call handler
	h := APIRootHandler{APIRootService: &ds}
	status, result := handlerTest(h.Get, "GET", testAPIRootURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var discovery cabby.APIRoot
	err := json.Unmarshal([]byte(result), &discovery)
	if err != nil {
		t.Fatal(err)
	}
	expected := testAPIRoot()

	if discovery.Title != expected.Title {
		t.Error("Got:", discovery.Title, "Expected:", expected.Title)
	}
	if discovery.Description != expected.Description {
		t.Error("Got:", discovery.Description, "Expected:", expected.Description)
	}
}

func TestAPIRootGetFailures(t *testing.T) {
	tests := []struct {
		method         string
		hasAPIRootErr  bool
		errorDesc      string
		expectedError  cabby.Error
		expectedStatus int
	}{
		{method: "GET",
			hasAPIRootErr:  true,
			errorDesc:      "APIRoot failure",
			expectedError:  cabby.Error{Title: "Internal Server Error"},
			expectedStatus: http.StatusInternalServerError},
	}

	for _, test := range tests {
		// finish setting up test
		test.expectedError.Description = test.errorDesc
		test.expectedError.HTTPStatus = test.expectedStatus

		ds := APIRootService{}
		ds.APIRootFn = func(path string) (cabby.APIRoot, error) {
			var err error
			if test.hasAPIRootErr {
				err = errors.New(test.errorDesc)
			}
			return cabby.APIRoot{}, err
		}

		h := APIRootHandler{APIRootService: &ds}
		status, result := handlerTest(h.Get, test.method, testAPIRootURL, nil)

		if status != test.expectedStatus {
			t.Error("Got:", status, "Expected:", test.expectedStatus)
		}

		var cabbyError cabby.Error
		err := json.Unmarshal([]byte(result), &cabbyError)
		if err != nil {
			t.Fatal(err)
		}

		if cabbyError.Title != test.expectedError.Title {
			t.Error("Got:", cabbyError.Title, "Expected:", test.expectedError.Title)
		}
		if cabbyError.Description != test.expectedError.Description {
			t.Error("Got:", cabbyError.Description, "Expected:", test.expectedError.Description)
		}
		if cabbyError.HTTPStatus != test.expectedError.HTTPStatus {
			t.Error("Got:", cabbyError.HTTPStatus, "Expected:", test.expectedError.HTTPStatus)
		}
	}
}

func TestAPIRootHandlerNoAPIRoot(t *testing.T) {
	ds := APIRootService{}
	ds.APIRootFn = func(path string) (cabby.APIRoot, error) {
		return cabby.APIRoot{Title: ""}, nil
	}

	h := APIRootHandler{APIRootService: &ds}
	status, result := handlerTest(h.Get, "GET", testAPIRootURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var cabbyError cabby.Error
	err := json.Unmarshal([]byte(result), &cabbyError)
	if err != nil {
		t.Fatal(err)
	}
	expected := cabby.Error{Title: "Resource not found",
		Description: "API Root not found: cabby_test_root", HTTPStatus: http.StatusNotFound}

	if cabbyError.Title != expected.Title {
		t.Error("Got:", cabbyError.Title, "Expected:", expected.Title)
	}
	if cabbyError.Description != expected.Description {
		t.Error("Got:", cabbyError.Description, "Expected:", expected.Description)
	}
	if cabbyError.HTTPStatus != expected.HTTPStatus {
		t.Error("Got:", cabbyError.HTTPStatus, "Expected:", expected.HTTPStatus)
	}
}
