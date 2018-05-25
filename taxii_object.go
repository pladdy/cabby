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

	result, err := sos.read(ts, tf)
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

	status, err := newTaxiiStatus()
	if err != nil {
		resourceNotFound(w, errors.New("Unable to process status"))
	}
	status.TotalCount = int64(len(bundle.Objects))

	// write status and close connection
	w.Header().Set("Connection", "close")
	writeContent(w, taxiiContentType, resourceToJSON(status))
	writeBundle(bundle, takeCollectionID(r), ts)
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

func writeBundle(b s.Bundle, cid string, ts taxiiStorer) {
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

	// is this dumb?  errors are logged in the taxiiStorer...what's the point of having them here?
	// ie: do i need to be passing an error channel around?
	for e := range writeErrs {
		log.Error(e)
	}
}
