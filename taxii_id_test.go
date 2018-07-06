package main

import "testing"

func TestNewTaxiiID(t *testing.T) {
	_, err := newTaxiiID()
	if err != nil {
		t.Error("Expected no error:", err)
	}
}

func TestTaxiiIDFromString(t *testing.T) {
	uid := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	id, err := taxiiIDFromString(uid)
	if err != nil {
		t.Error("Expected no error:", err)
	}

	if id.String() != uid {
		t.Error("Got:", id.String(), "Expected:", uid)
	}
}

func TestTaxiiIDFromStringBadInput(t *testing.T) {
	_, err := taxiiIDFromString("")
	if err == nil {
		t.Error("Expected an error")
	}

	_, err = taxiiIDFromString("fail")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestTaxiiIDUsingStringBadInput(t *testing.T) {
	_, err := taxiiIDFromString("")
	if err == nil {
		t.Error("Expected an error")
	}

	_, err = taxiiIDFromString("fail")
	if err == nil {
		t.Error("Expected an error")
	}
}

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
