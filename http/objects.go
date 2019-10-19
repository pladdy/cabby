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
	w.Header().Set("Allow", VersionsMethods)
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}

// Get handles a get request for the objects endpoint
func (h ObjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ObjectsHandler"}).Debug("Handler called")

	if !verifySupportedMimeType(w, r, "Accept", cabby.TaxiiContentType) {
		return
	}

	h.getObjects(w, r)
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

	envelope := objectsToEnvelope(objects, cr)

	w.Header().Set("X-TAXII-Date-Added-First", cr.AddedAfterFirst())
	w.Header().Set("X-TAXII-Date-Added-Last", cr.AddedAfterLast())

	if cr.Set {
		w.Header().Set("Content-Range", cr.String())
		writePartialContent(w, r, cabby.TaxiiContentType, resourceToJSON(envelope))
	} else {
		writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(envelope))
	}
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
	if !verifySupportedMimeType(w, r, "Accept", cabby.TaxiiContentType) {
		return
	}

	if !verifySupportedMimeType(w, r, "Content-Type", cabby.TaxiiContentType) {
		return
	}

	if r.ContentLength > h.MaxContentLength {
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
