package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

/* api root handlers */

func handleAdminTaxiiAPIRoot(ts taxiiStorer, h *http.ServeMux) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		if !userIsAuthorized(w, r) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			badRequest(w, err)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			handleAdminTaxiiAPIRootDelete(ts, w, body)
		case http.MethodPost:
			handleAdminTaxiiAPIRootPost(ts, h, w, body)
		case http.MethodPut:
			handleAdminTaxiiAPIRootPut(ts, w, body)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}
	})
}

func handleAdminTaxiiAPIRootDelete(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	tar, err := bodyToAPIRoot(body)
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

func handleAdminTaxiiAPIRootPost(ts taxiiStorer, handler *http.ServeMux, w http.ResponseWriter, body []byte) {
	tar, err := bodyToAPIRoot(body)
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

func handleAdminTaxiiAPIRootPut(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	tar, err := bodyToAPIRoot(body)
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

		if !userIsAuthorized(w, r) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			badRequest(w, err)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			handleAdminTaxiiCollectionsDelete(ts, w, body)
		case http.MethodPost:
			handleAdminTaxiiCollectionsPost(ts, w, body, takeUser(r))
		case http.MethodPut:
			handleAdminTaxiiCollectionsPut(ts, w, body)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}
	})
}

func handleAdminTaxiiCollectionsDelete(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	tc, err := bodyToCollection(body)
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

func handleAdminTaxiiCollectionsPost(ts taxiiStorer, w http.ResponseWriter, body []byte, user string) {
	tc, err := bodyToCollection(body)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tc.create(ts, user, tc.APIRootPath)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(tc))
}

func handleAdminTaxiiCollectionsPut(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	tc, err := bodyToCollection(body)
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

		if !userIsAuthorized(w, r) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			badRequest(w, err)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			handleAdminTaxiiDiscoveryDelete(ts, w)
		case http.MethodPost:
			handleAdminTaxiiDiscoveryPost(ts, w, body)
		case http.MethodPut:
			handleAdminTaxiiDiscoveryPut(ts, w, body)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}
	})
}

func handleAdminTaxiiDiscoveryDelete(ts taxiiStorer, w http.ResponseWriter) {
	td := taxiiDiscovery{}
	err := td.delete(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, jsonContentType, `{"deleted": "discovery"}`)
}

func handleAdminTaxiiDiscoveryPost(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	td, err := bodyToDiscovery(body)
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

func handleAdminTaxiiDiscoveryPut(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	td, err := bodyToDiscovery(body)
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

/* user */

func handleAdminTaxiiUser(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		if !userIsAuthorized(w, r) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			badRequest(w, err)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			handleAdminTaxiiUserDelete(ts, w, body)
		case http.MethodPost:
			handleAdminTaxiiUserPost(ts, w, body)
		case http.MethodPut:
			handleAdminTaxiiUserPut(ts, w, body)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}
	})
}

func handleAdminTaxiiUserDelete(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	tu, err := bodyToUser(body)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tu.delete(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, jsonContentType, `{"deleted": "`+tu.Email+`"}`)
}

func handleAdminTaxiiUserPost(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	// create the user
	tu, err := bodyToUser(body)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tu.create(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	// create the password for the user
	tup, err := bodyToUserPassword(body)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tup.create(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(tu))
}

func handleAdminTaxiiUserPut(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	tu, err := bodyToUser(body)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tu.update(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(tu))
}

/* user collection */

func handleAdminTaxiiUserCollection(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		if !userIsAuthorized(w, r) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			badRequest(w, err)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			handleAdminTaxiiUserCollectionDelete(ts, w, body)
		case http.MethodPost:
			handleAdminTaxiiUserCollectionPost(ts, w, body)
		case http.MethodPut:
			handleAdminTaxiiUserCollectionPut(ts, w, body)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}
	})
}

func handleAdminTaxiiUserCollectionDelete(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	tuc, err := bodyToUserCollection(body)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tuc.delete(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, jsonContentType, `{"deleted": "`+tuc.taxiiCollectionAccess.ID.String()+`"}`)
}

func handleAdminTaxiiUserCollectionPost(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	tuc, err := bodyToUserCollection(body)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tuc.create(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(tuc))
}

func handleAdminTaxiiUserCollectionPut(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	tuc, err := bodyToUserCollection(body)
	fmt.Println(tuc)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tuc.update(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(tuc))
}

/* user password */

func handleAdminTaxiiUserPassword(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		if !userIsAuthorized(w, r) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			badRequest(w, err)
			return
		}

		switch r.Method {
		case http.MethodPut:
			handleAdminTaxiiUserPasswordPut(ts, w, body)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}
	})
}

func handleAdminTaxiiUserPasswordPut(ts taxiiStorer, w http.ResponseWriter, body []byte) {
	tup, err := bodyToUserPassword(body)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tup.update(ts)
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, `{"updated": "`+tup.Email+`"}`)
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

func bodyToAPIRoot(b []byte) (taxiiAPIRoot, error) {
	var resource taxiiAPIRoot
	err := json.Unmarshal(b, &resource)
	return resource, err
}

func bodyToCollection(b []byte) (taxiiCollection, error) {
	var resource taxiiCollection
	err := json.Unmarshal(b, &resource)
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

func bodyToDiscovery(b []byte) (taxiiDiscovery, error) {
	var resource taxiiDiscovery
	err := json.Unmarshal(b, &resource)
	return resource, err
}

func bodyToUser(b []byte) (taxiiUser, error) {
	var resource taxiiUser
	err := json.Unmarshal(b, &resource)
	return resource, err
}

func bodyToUserCollection(b []byte) (taxiiUserCollection, error) {
	var resource taxiiUserCollection
	err := json.Unmarshal(b, &resource)
	return resource, err
}

func bodyToUserPassword(b []byte) (taxiiUserPassword, error) {
	var resource taxiiUserPassword
	err := json.Unmarshal(b, &resource)
	return resource, err
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
