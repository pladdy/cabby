package sqlite

import (
	"testing"
)

func TestNewDataStore(t *testing.T) {
	_, err := NewDataStore("temp.db")

	if err != nil {
		t.Error("Got:", err, "Expected: nil")
	}
}

func TestNewDataStoreNoPath(t *testing.T) {
	_, err := NewDataStore("")

	if err == nil {
		t.Error("Expected an error")
	}
}

func TestDataStoreClose(t *testing.T) {
	s, err := NewDataStore("temp.db")

	if err != nil {
		t.Error("Got:", err, "Expected: nil")
	}

	s.Close()
}
