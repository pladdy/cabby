package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

/* handlers */

func handleAdminTaxiiAPIRoot(ts taxiiStorer, h *http.ServeMux) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		switch r.Method {
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

func handleAdminTaxiiAPIRootPost(ts taxiiStorer, handler *http.ServeMux, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	tar, err := bodyToAPIRoot(w, r)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tar.create(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	registerAPIRoot(ts, tar.Path, handler)
	writeContent(w, taxiiContentType, resourceToJSON(tar))
}

func handleAdminTaxiiAPIRootPut(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	tar, err := bodyToAPIRoot(w, r)
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

func bodyToAPIRoot(w http.ResponseWriter, r *http.Request) (taxiiAPIRoot, error) {
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
