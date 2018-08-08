package http

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	cabby "github.com/pladdy/cabby2"
)

func TestErrorStatus(t *testing.T) {
	res := httptest.NewRecorder()

	expectedTitle := "A test"
	expectedDescription := "fake error"
	expectedStatus := http.StatusInternalServerError

	errorStatus(res, expectedTitle, errors.New(expectedDescription), expectedStatus)
	body, _ := ioutil.ReadAll(res.Body)

	var result cabby.Error
	err := json.Unmarshal(body, &result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Title != expectedTitle {
		t.Error("Got:", result.Title, "Expected:", expectedTitle)
	}
	if result.Description != expectedDescription {
		t.Error("Got:", result.Description, "Expected:", expectedDescription)
	}
	if result.HTTPStatus != expectedStatus {
		t.Error("Got:", result.HTTPStatus, "Expected:", expectedStatus)
	}
}

func TestRecoverFromPanic(t *testing.T) {
	w := httptest.NewRecorder()
	defer recoverFromPanic(w)
	panic("test")
}
