package main

import (
	"bytes"
	"net/http"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestRegisterAPIRootInvalidPath(t *testing.T) {
	setupSQLite()

	// remove required table
	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_api_root")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	invalidPath := "foo"
	handler := http.NewServeMux()
	registerAPIRoot(ts, invalidPath, handler)
}

func TestRegisterCollectionRoutesFail(t *testing.T) {
	setupSQLite()

	var buf bytes.Buffer

	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// remove required table
	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	handler := http.NewServeMux()
	registerCollectionRoutes(ts, taxiiAPIRoot{}, "test", handler)

	if len(buf.String()) == 0 {
		t.Error("Expected log output")
	}
}
