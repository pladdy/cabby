package main

import "testing"

func TestTaxiiIDIsEmpty(t *testing.T) {
	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	if id.isEmpty() == true {
		t.Error("Expected to NOT be empty")
	}

	emptyID := taxiiID{}
	if emptyID.isEmpty() == false {
		t.Error("Expected ID to be empty")
	}
}
