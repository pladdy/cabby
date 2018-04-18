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

	b, err := s.NewBundle()
	if err != nil {
		resourceNotFound(w, errors.New("Unable to create bundle"))
		return
	}

	stixID := getStixID(r.URL.Path)
	collectionID := getCollectionID(r.URL.Path)

	sos, err := getStixObjects(ts, collectionID, stixID)
	if err != nil {
		log.WithFields(
			log.Fields{"fn": "handleGetTaxiiObjects", "error": err, "stixID": stixID, "collection": collectionID},
		).Error("failed to get objects")
		resourceNotFound(w, errors.New("Unable to process request"))
		return
	}

	for _, o := range sos.Objects {
		b.Objects = append(b.Objects, o)
	}
	writeContent(w, stixContentType, resourceToJSON(b))
}

func getStixObjects(ts taxiiStorer, collectionID, stixID string) (stixObjects, error) {
	sos := stixObjects{}
	var err error

	if len(stixID) > 0 {
		err = sos.read(ts, collectionID, stixID)
	} else {
		err = sos.read(ts, collectionID)
	}
	return sos, err
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
	writeContent(w, taxiiContentType, resourceToJSON(status))
	go writeBundle(bundle, getCollectionID(r.URL.Path), ts)
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
