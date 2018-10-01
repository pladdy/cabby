package http

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestErrorStatus(t *testing.T) {
	res := httptest.NewRecorder()

	expected := cabby.Error{Title: "A test",
		Description: "fake error", HTTPStatus: http.StatusInternalServerError}

	errorStatus(res, expected.Title, errors.New(expected.Description), expected.HTTPStatus)
	body, _ := ioutil.ReadAll(res.Body)

	var result cabby.Error
	err := json.Unmarshal(body, &result)
	if err != nil {
		t.Fatal(err)
	}

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestRecoverFromPanic(t *testing.T) {
	w := httptest.NewRecorder()
	defer recoverFromPanic(w)
	panic("test")
}
