// each CRUD interface has a matching function type i'm using for logging decorators.  this feels gross but i don't know
// a better way.  also i thought it would be worse to try and log in each method within the backend implementation.  at
// least this way the decoratoers are defined here and used in the data store implementations...

package main

import (
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
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

type taxiiCreator interface {
	create(resource string, toWrite chan interface{}, errors chan error)
}

type taxiiCreatorFunc func(resource string, toWrite chan interface{}, errors chan error)

type taxiiDeleter interface {
	delete(resource string, args []interface{}) error
}

type taxiiDeleterFunc func(resource string, args []interface{}) error

type taxiiQuery struct {
	resource  string
	statement string
}

type taxiiReader interface {
	read(resource string, args []interface{}, tf ...taxiiFilter) (taxiiResult, error)
}

type taxiiReaderFunc func(resource string, args []interface{}, tf ...taxiiFilter) (taxiiResult, error)

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

type taxiiUpdaterFunc func(resource string, args []interface{}) error

type taxiiStorer interface {
	taxiiConnector
	taxiiDeleter
	taxiiReader
	taxiiCreator
	taxiiUpdater
}

/* helpers */

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

func logResourceStart(resource, action string) time.Time {
	milliSecondOfNanoSeconds := int64(1000000)

	start := time.Now().In(time.UTC)
	log.WithFields(log.Fields{
		"action":   action,
		"resource": resource,
		"start_ts": start.UnixNano() / milliSecondOfNanoSeconds,
	}).Info("Starting with resource")
	return start
}

func logResourceEnd(resource, action string, start time.Time) {
	milliSecondOfNanoSeconds := int64(1000000)

	end := time.Now().In(time.UTC)
	elapsed := time.Since(start)

	log.WithFields(log.Fields{
		"action":     action,
		"elapsed_ts": float64(elapsed.Nanoseconds()) / float64(milliSecondOfNanoSeconds),
		"end_ts":     end.UnixNano() / milliSecondOfNanoSeconds,
		"resource":   resource,
	}).Info("Finished with resource")
}

func newTaxiiStorer(ds, path string) (t taxiiStorer, err error) {
	if ds == "sqlite" {
		t, err = newSQLiteDB(path)
	} else {
		err = errors.New("Unsupported data store specified in config")
	}
	return
}

func withCreatorLogging(tcf taxiiCreatorFunc) taxiiCreatorFunc {
	return func(resource string, toWrite chan interface{}, errors chan error) {
		action := "create"
		start := logResourceStart(resource, action)
		tcf(resource, toWrite, errors)
		logResourceEnd(resource, action, start)
	}
}

func withDeleterLogging(tdf taxiiDeleterFunc) taxiiDeleterFunc {
	return func(resource string, args []interface{}) error {
		action := "delete"
		start := logResourceStart(resource, action)
		err := tdf(resource, args)
		logResourceEnd(resource, action, start)
		return err
	}
}

func withReaderLogging(trf taxiiReaderFunc) taxiiReaderFunc {
	return func(resource string, args []interface{}, tf ...taxiiFilter) (taxiiResult, error) {
		action := "read"
		start := logResourceStart(resource, action)
		result, err := trf(resource, args, tf...)
		logResourceEnd(resource, action, start)
		return result, err
	}
}

func withUpdaterLogging(tuf taxiiUpdaterFunc) taxiiUpdaterFunc {
	return func(resource string, args []interface{}) error {
		action := "update"
		start := logResourceStart(resource, action)
		err := tuf(resource, args)
		logResourceEnd(resource, action, start)
		return err
	}
}
