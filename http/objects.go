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
const ObjectsMethods = "Get, Head, Post"

// ObjectMethods lists allowed methods
const ObjectMethods = "Get, Delete, Head"

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

	err := h.ObjectService.DeleteObject(r.Context(), takeCollectionID(r), takeObjectID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}
	writeContent(w, r, cabby.TaxiiContentType, "")
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

	envelope, err := objectsToEnvelope(objects)
	if err != nil {
		internalServerError(w, errors.New("Unable to create envelope"))
		return
	}

	w.Header().Set("X-TAXII-Date-Added-First", cr.AddedAfterFirst())
	w.Header().Set("X-TAXII-Date-Added-Last", cr.AddedAfterLast())

	if cr.Set {
		w.Header().Set("Content-Range", cr.String())
		writePartialContent(w, r, cabby.StixContentType, resourceToJSON(envelope))
	} else {
		writeContent(w, r, cabby.StixContentType, resourceToJSON(envelope))
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

	envelope, err := objectsToEnvelope(objects)
	if err != nil {
		internalServerError(w, errors.New("Unable to create envelope"))
		return
	}

	writeContent(w, r, cabby.StixContentType, resourceToJSON(envelope))
}

/* Post */

// Post handles post request
func (h ObjectsHandler) Post(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ObjectsHandler"}).Debug("Handler called")

	// due to how the routing logic works, a post can go to a url with an object on the path.  a path with
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

	envelope, err := envelopeFromBytes(body)
	if err != nil || len(envelope.Objects) == 0 {
		badRequest(w, err)
		return
	}

	status, err := cabby.NewStatus(len(envelope.Objects))
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

	go h.ObjectService.CreateEnvelope(r.Context(), envelope, takeCollectionID(r), status, h.StatusService)
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

func envelopeFromBytes(b []byte) (e cabby.Envelope, err error) {
	err = json.Unmarshal(b, &e)
	if err != nil {
		log.WithFields(log.Fields{"envelope": string(b), "error:": err}).Error("Unable to unmarshal envelope")
		return e, fmt.Errorf("Unable to convert JSON to envelope, error: %v", err)
	}
	return
}

func greaterThan(r, m int64) bool {
	if r > m {
		return true
	}
	return false
}

func handlePostToObjectURL(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"object id": takeObjectID(r)}).Error("Invalid method for object")
	w.Header().Set("Allow", ObjectMethods)
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}

func objectsToEnvelope(objects []stones.Object) (e cabby.Envelope, err error) {
	for _, o := range objects {
		e.Objects = append(e.Objects, json.RawMessage(o.Source))
	}

	if len(e.Objects) == 0 {
		log.Warn("Can't return an empty envelope, returning error to caller")
		return e, errors.New("No data returned: empty envelope")
	}
	return
}
