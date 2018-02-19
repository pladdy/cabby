package main

import (
	"testing"
)

func TestTaxiiStorer(t *testing.T) {
	ts, err := newTaxiiStorer(config.DataStore["name"], config.DataStore["path"])
	if err != nil {
		t.Error(err)
	}
	defer ts.disconnect()
}

func TestTaxiiStorerFail(t *testing.T) {
	_, err := newTaxiiStorer("no store", config.DataStore["path"])

	if err == nil {
		t.Error("Expected error")
	}
}
