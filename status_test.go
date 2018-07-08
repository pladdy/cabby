package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleUndefinedRequest(t *testing.T) {
	status, result := handlerTest(handleUndefinedRequest, "GET", "/nobody-home", nil)
	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound, "Response:", result)
	}
}

func TestRecoverFromPanic(t *testing.T) {
	w := httptest.NewRecorder()
	defer recoverFromPanic(w)
	panic("test")
}
