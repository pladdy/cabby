package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	s "github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

func handleTaxiiObjects(ts taxiiStorer, maxContentLength int64) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		switch r.Method {
		case http.MethodPost:
			if contentTooLarge(r.ContentLength, maxContentLength) {
				requestTooLarge(w, r.ContentLength, maxContentLength)
				return
			}
			handlePostTaxiiObjects(ts, w, r)
		case http.MethodGet:
			handleGetTaxiiObjects(ts, w, r)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
		}
	})
}

func handleGetTaxiiObjects(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !takeCollectionAccess(r).CanRead {
		unauthorized(w, fmt.Errorf("Unauthorized to read from collection"))
		return
	}

	tr, err := newTaxiiRange(r.Header.Get("Range"))
	if err != nil {
		rangeNotSatisfiable(w, err)
		return
	}
	r = withTaxiiRange(r, tr)

	result, err := getStixObjects(ts, r)
	if err != nil {
		resourceNotFound(w, errors.New("Unable to process request"))
		return
	}

	b, err := stixObjectsToBundle(result.data.(stixObjects))
	if err != nil {
		resourceNotFound(w, errors.New("Unable to create bundle"))
		return
	}

	if tr.Valid() {
		tr.total = result.items
		w.Header().Set("Content-Range", tr.String())
		writePartialContent(w, stixContentType, resourceToJSON(b))
	} else {
		writeContent(w, stixContentType, resourceToJSON(b))
	}
}

func getStixObjects(ts taxiiStorer, r *http.Request) (taxiiResult, error) {
	tf := newTaxiiFilter(r)
	sos := stixObjects{}

	result, err := sos.read(ts, tf, takeObjectID(r))
	if err != nil {
		log.WithFields(
			log.Fields{"fn": "getStixObjects", "error": err, "taxiiFilter": tf},
		).Error("failed to get objects")
	}
	return result, err
}

func handlePostTaxiiObjects(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !takeCollectionAccess(r).CanWrite {
		unauthorized(w, fmt.Errorf("Unauthorized to write to collection"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		badRequest(w, err)
		return
	}
	defer r.Body.Close()

	bundle, err := bundleFromBytes(body)
	if err != nil {
		badRequest(w, err)
		return
	}

	status, err := newTaxiiStatus(len(bundle.Objects))
	if err != nil {
		internalServerError(w, errors.New("Unable to create status resource"))
	}

	err = status.create(ts)
	if err != nil {
		internalServerError(w, errors.New("Unable to store status resource"))
	}

	w.WriteHeader(http.StatusAccepted)
	writeContent(w, taxiiContentType, resourceToJSON(status))
	go writeBundle(bundle, takeCollectionID(r), ts, status)
}

/* helpers */

func bundleFromBytes(b []byte) (s.Bundle, error) {
	var bundle s.Bundle

	err := json.Unmarshal(b, &bundle)
	if err != nil {
		return bundle, fmt.Errorf("Unable to convert json to bundle, error: %v", err)
	}

	valid, errs := bundle.Validate()
	if !valid {
		errString := "Invalid bundle:"
		for _, e := range errs {
			errString = errString + fmt.Sprintf("\nValidation Error: %v", e)
		}
		err = fmt.Errorf(errString)
	}

	return bundle, err
}

func contentTooLarge(r, m int64) bool {
	if r > m {
		return true
	}
	return false
}

func updateStatus(errs chan error, s taxiiStatus, ts taxiiStorer) {
	failures := int64(0)
	for _ = range errs {
		failures++
	}

	s.FailureCount = failures
	s.SuccessCount = s.TotalCount - failures
	s.PendingCount = s.TotalCount - s.SuccessCount - failures

	s.update(ts, "complete")
}

func writeBundle(b s.Bundle, cid string, ts taxiiStorer, s taxiiStatus) {
	writeErrs := make(chan error, len(b.Objects))
	writes := make(chan interface{}, minBuffer)

	go ts.create("stixObject", writes, writeErrs)

	for _, object := range b.Objects {
		so, err := bytesToStixObject(object)
		if err != nil {
			writeErrs <- err
			continue
		}
		log.WithFields(log.Fields{"stix_id": so.RawID}).Info("Sending to data store")
		writes <- []interface{}{so.RawID, so.Type, so.Created, so.Modified, so.Object, cid}
	}

	close(writes)
	updateStatus(writeErrs, s, ts)
}
