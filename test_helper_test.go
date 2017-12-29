// test helper file to declare top level vars/constants and define helper functions for all tests

package main

import (
	"log"
	"os"
	"testing"
)

var sqlDriver = "sqlite3"
var testDB = "test/test.db"

/* helpers */

func renameFile(from, to string) {
	err := os.Rename(from, to)
	if err != nil {
		log.Fatal("Failed to rename file:", from, "to:", to)
	}
}

/* check for panics */

type panicChecker struct {
	recovered bool
}

func attemptRecover(t *testing.T, p *panicChecker) {
	if err := recover(); err == nil {
		t.Error("Failed to recover:", err)
	}
	p.recovered = true
}
