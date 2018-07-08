package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

/* handlers */

func handleAdminTaxiiAPIRoot(ts taxiiStorer, h *http.ServeMux) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		switch r.Method {
		case http.MethodDelete:
			handleAdminTaxiiAPIRootDelete(ts, w, r)
		case http.MethodPost:
			handleAdminTaxiiAPIRootPost(ts, h, w, r)
		case http.MethodPut:
			handleAdminTaxiiAPIRootPut(ts, w, r)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}
	})
}

func handleAdminTaxiiAPIRootDelete(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	tar, err := bodyToAPIRoot(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tar.delete(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, jsonContentType, `{"deleted": "`+tar.Path+`"}`)
}

func handleAdminTaxiiAPIRootPost(ts taxiiStorer, handler *http.ServeMux, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	tar, err := bodyToAPIRoot(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tar.create(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	attemptRegisterAPIRoot(ts, tar.Path, handler)
	writeContent(w, taxiiContentType, resourceToJSON(tar))
}

func handleAdminTaxiiAPIRootPut(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	tar, err := bodyToAPIRoot(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tar.update(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(tar))
}

/* helpers */

func attemptRegisterAPIRoot(ts taxiiStorer, path string, s *http.ServeMux) {
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{"path": path}).Warn("failed to register api root handlers; could already be registered")
		}
	}()

	registerAPIRoot(ts, path, s)
}

func bodyToAPIRoot(r *http.Request) (taxiiAPIRoot, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return taxiiAPIRoot{}, err
	}
	defer r.Body.Close()

	var tar taxiiAPIRoot

	err = json.Unmarshal(body, &tar)
	return tar, err
}

func userIsAuthorized(w http.ResponseWriter, r *http.Request) bool {
	if !takeUser(r) {
		unauthorized(w, errors.New("No user specified"))
		return false
	}

	if !takeCanAdmin(r) {
		forbidden(w, errors.New("Not authorized to create API Roots"))
		return false
	}
	return true
}
