package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// ObjectsMethods lists allowed methods
const ObjectsMethods = "Get, Head, Post"

// ObjectsHandler handles Objects requests
type ObjectsHandler struct {
	ObjectService    cabby.ObjectService
	StatusService    cabby.StatusService
	MaxContentLength int64
}

// Delete handler
func (h ObjectsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, r, ObjectsMethods)
}

// Get handles a get request for the objects endpoint
func (h ObjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ObjectsHandler"}).Debug("Handler called")

	if !verifyRequestHeader(r, "Accept", cabby.TaxiiContentType) {
		notAcceptable(w, fmt.Errorf("Accept header must be '%v'", cabby.TaxiiContentType))
		return
	}

	if !requestIsReadAuthorized(r) {
		forbidden(w, errors.New("Unauthorized access"))
		return
	}

	p, err := cabby.NewPage(takeLimit(r))
	if err != nil {
		badRequest(w, err)
		return
	}

	objects, err := h.ObjectService.Objects(r.Context(), takeCollectionID(r), &p, newFilter(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if noResources(len(objects)) {
		resourceNotFound(w, errors.New("No resources available for this request"))
		return
	}

	envelope := objectsToEnvelope(objects, p)
	w.Header().Set("X-TAXII-Date-Added-First", p.AddedAfterFirst())
	w.Header().Set("X-TAXII-Date-Added-Last", p.AddedAfterLast())
	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(envelope))
}

/* Post */

// Post handles post request
func (h ObjectsHandler) Post(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ObjectsHandler"}).Debug("Handler called")

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
	if !verifyRequestHeader(r, "Accept", cabby.TaxiiContentType) {
		notAcceptable(w, fmt.Errorf("Accept header must be '%v'", cabby.TaxiiContentType))
		return
	}

	if !verifyRequestHeader(r, "Content-Type", cabby.TaxiiContentType) {
		unsupportedMediaType(w, fmt.Errorf("Content-Type header must be '%v'", cabby.TaxiiContentType))
		return
	}

	if r.ContentLength > h.MaxContentLength {
		requestTooLarge(w, r.ContentLength, h.MaxContentLength)
		return
	}

	if !requestIsWriteAuthorized(r) {
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
