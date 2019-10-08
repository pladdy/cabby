package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

// ObjectsMethods lists allowed methods
const ObjectsMethods = "Get, Head"

// ObjectsHandler handles Objects requests
type ObjectsHandler struct {
	ObjectService    cabby.ObjectService
	StatusService    cabby.StatusService
	MaxContentLength int64
}

/* Delete */

//Delete handles a delete of an object; can only be done given an ID
func (h ObjectsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ObjectsHandler"}).Debug("Handler called")

	if takeObjectID(r) == "" {
		w.Header().Set("Allow", ObjectsMethods)
		methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
		return
	}
	h.deleteObject(w, r)
}

func (h ObjectsHandler) deleteObject(w http.ResponseWriter, r *http.Request) {
	// implement
	return
}

/* Get */

// Get handles a get request for the objects endpoint; it has to decide if a request is for objects in a collection
// or a specfic object
func (h ObjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ObjectsHandler"}).Debug("Handler called")

	if !verifySupportedMimeType(w, r, "Accept", cabby.StixContentType) {
		return
	}

	if takeObjectID(r) == "" {
		h.getObjects(w, r)
		return
	}
	h.getObject(w, r)
}

func (h ObjectsHandler) getObjects(w http.ResponseWriter, r *http.Request) {
	cr, err := cabby.NewRange(r.Header.Get("Range"))
	if err != nil {
		rangeNotSatisfiable(w, err, cr)
		return
	}

	objects, err := h.ObjectService.Objects(r.Context(), takeCollectionID(r), &cr, newFilter(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if noResources(w, len(objects), cr) {
		return
	}

	bundle, err := objectsToBundle(objects)
	if err != nil {
		internalServerError(w, errors.New("Unable to create bundle"))
		return
	}

	w.Header().Set("X-TAXII-Date-Added-First", cr.AddedAfterFirst())
	w.Header().Set("X-TAXII-Date-Added-Last", cr.AddedAfterLast())

	if cr.Set {
		w.Header().Set("Content-Range", cr.String())
		writePartialContent(w, r, cabby.StixContentType, resourceToJSON(bundle))
	} else {
		writeContent(w, r, cabby.StixContentType, resourceToJSON(bundle))
	}
}

func (h ObjectsHandler) getObject(w http.ResponseWriter, r *http.Request) {
	objects, err := h.ObjectService.Object(r.Context(), takeCollectionID(r), takeObjectID(r), newFilter(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if len(objects) <= 0 {
		resourceNotFound(w, errors.New("No objects defined in this collection"))
		return
	}

	bundle, err := objectsToBundle(objects)
	if err != nil {
		internalServerError(w, errors.New("Unable to create bundle"))
		return
	}

	writeContent(w, r, cabby.StixContentType, resourceToJSON(bundle))
}

/* Post */

// Post handles post request
func (h ObjectsHandler) Post(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ObjectsHandler"}).Debug("Handler called")

	// due to how the routing logic works, a post can go to a url with an object on the path.  technically a path with
	// an object id can only receive a get method
	if takeObjectID(r) != "" {
		handlePostToObjectURL(w, r)
		return
	}

	if !h.validPost(w, r) {
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

	status, err := cabby.NewStatus(len(bundle.Objects))
	if err != nil {
		internalServerError(w, errors.New("Unable to initialize status resource"))
		return
	}

	err = h.StatusService.CreateStatus(r.Context(), status)
	if err != nil {
		internalServerError(w, errors.New("Unable to store status resource"))
		return
	}

	// write header before status or header won't be set
	w.Header().Set("Content-Type", cabby.TaxiiContentType)
	w.WriteHeader(http.StatusAccepted)
	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(status))

	go h.ObjectService.CreateBundle(r.Context(), bundle, takeCollectionID(r), status, h.StatusService)
}

func (h ObjectsHandler) validPost(w http.ResponseWriter, r *http.Request) (isValid bool) {
	if !verifySupportedMimeType(w, r, "Accept", cabby.TaxiiContentType) {
		return
	}

	if !verifySupportedMimeType(w, r, "Content-Type", cabby.StixContentType) {
		return
	}

	if greaterThan(r.ContentLength, h.MaxContentLength) {
		requestTooLarge(w, r.ContentLength, h.MaxContentLength)
		return
	}

	if !takeCollectionAccess(r).CanWrite {
		forbidden(w, fmt.Errorf("Unauthorized to write to collection"))
		return
	}

	return true
}

/* helpers */

func bundleFromBytes(b []byte) (stones.Bundle, error) {
	var bundle stones.Bundle

	err := json.Unmarshal(b, &bundle)
	if err != nil {
		log.WithFields(log.Fields{"bundle": string(b), "error:": err}).Error("Unable to unmarshal bundle")
		return bundle, fmt.Errorf("Unable to convert JSON to bundle, error: %v", err)
	}

	valid, errs := bundle.Valid()
	if !valid {
		return bundle, stones.ErrorsToString(errs)
	}
	return bundle, nil
}

func greaterThan(r, m int64) bool {
	if r > m {
		return true
	}
	return false
}

func handlePostToObjectURL(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"object id": takeObjectID(r)}).Error("Invalid method for object")
	w.Header().Set("Allow", ObjectsMethods)
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}

func objectsToBundle(objects []stones.Object) (stones.Bundle, error) {
	bundle, err := stones.NewBundle()
	if err != nil {
		return bundle, err
	}

	for _, o := range objects {
		bundle.AddObject(string(o.Source))
	}

	if len(bundle.Objects) == 0 {
		log.Warn("Can't return an empty bundle, returning error to caller")
		return bundle, errors.New("No data returned: empty bundle")
	}
	return bundle, err
}
