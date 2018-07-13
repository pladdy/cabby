package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

/* api root handlers */

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

/* collection handlers */

func handleAdminTaxiiCollections(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		switch r.Method {
		case http.MethodDelete:
			handleAdminTaxiiCollectionsDelete(ts, w, r)
		case http.MethodPost:
			handleAdminTaxiiCollectionsPost(ts, w, r)
		case http.MethodPut:
			handleAdminTaxiiCollectionsPut(ts, w, r)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}
	})
}

func handleAdminTaxiiCollectionsDelete(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	tc, err := bodyToCollection(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tc.delete(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, jsonContentType, `{"deleted": "`+tc.ID.String()+`"}`)
}

func handleAdminTaxiiCollectionsPost(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	tc, err := bodyToCollection(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tc.create(ts, takeUser(r), tc.APIRootPath)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(tc))
}

func handleAdminTaxiiCollectionsPut(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	tc, err := bodyToCollection(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tc.update(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(tc))
}

/* discovery handlers */

func handleAdminTaxiiDiscovery(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		switch r.Method {
		case http.MethodDelete:
			handleAdminTaxiiDiscoveryDelete(ts, w, r)
		case http.MethodPost:
			handleAdminTaxiiDiscoveryPost(ts, w, r)
		case http.MethodPut:
			handleAdminTaxiiDiscoveryPut(ts, w, r)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}
	})
}

func handleAdminTaxiiDiscoveryDelete(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	td := taxiiDiscovery{}
	err := td.delete(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, jsonContentType, `{"deleted": "discovery"}`)
}

func handleAdminTaxiiDiscoveryPost(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	td, err := bodyToDiscovery(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = td.create(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(td))
}

func handleAdminTaxiiDiscoveryPut(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	if !userIsAuthorized(w, r) {
		return
	}

	td, err := bodyToDiscovery(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = td.update(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(td))
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
	body, err := takeBody(r)
	var resource taxiiAPIRoot
	err = json.Unmarshal(body, &resource)
	return resource, err
}

func bodyToCollection(r *http.Request) (taxiiCollection, error) {
	body, err := takeBody(r)

	var resource taxiiCollection
	err = json.Unmarshal(body, &resource)
	if err != nil {
		return resource, err
	}

	err = resource.ensureID()
	if err != nil {
		return resource, err
	}
	resource.CanRead = true
	resource.CanWrite = true

	return resource, err
}

func bodyToDiscovery(r *http.Request) (taxiiDiscovery, error) {
	body, err := takeBody(r)
	var resource taxiiDiscovery
	err = json.Unmarshal(body, &resource)
	return resource, err
}

func takeBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	return body, err
}

func userIsAuthorized(w http.ResponseWriter, r *http.Request) bool {
	if !userExists(r) {
		unauthorized(w, errors.New("No user specified"))
		return false
	}

	if !takeCanAdmin(r) {
		forbidden(w, errors.New("Not authorized to create API Roots"))
		return false
	}
	return true
}
