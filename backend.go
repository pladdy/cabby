package main

import (
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

const (
	backendDir = "backend"
	maxWrites  = 500
	minBuffer  = 10
)

type taxiiConnector interface {
	connect(connection string) error
	disconnect()
}

type taxiiQuery struct {
	resource  string
	statement string
}

type taxiiReader interface {
	read(resource string, args []interface{}, tf ...taxiiFilter) (taxiiResult, error)
}

type taxiiResult struct {
	data         interface{}
	itemStart    int64
	itemEnd      int64
	items        int64
	query        taxiiQuery
	queryRunTime int64
}

func (t *taxiiResult) withPagination(tr taxiiRange) {
	result := *t
	result.itemStart = tr.first
	result.itemEnd = tr.last
	*t = result
}

type taxiiUpdater interface {
	update(resource string, args []interface{}) error
}

type taxiiWriter interface {
	create(resource string, toWrite chan interface{}, errors chan error)
}

type taxiiStorer interface {
	taxiiConnector
	taxiiReader
	taxiiWriter
	taxiiUpdater
}

/* helpers */

func newTaxiiStorer(ds, path string) (t taxiiStorer, err error) {
	if ds == "sqlite" {
		t, err = newSQLiteDB(path)
	} else {
		err = errors.New("Unsupported data store specified in config")
	}
	return
}

func createResource(ts taxiiStorer, resource string, args []interface{}) error {
	var err error

	toWrite := make(chan interface{}, minBuffer)
	errs := make(chan error, minBuffer)

	go ts.create(resource, toWrite, errs)
	toWrite <- args
	close(toWrite)

	for e := range errs {
		err = e
	}

	return err
}
