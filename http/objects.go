package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

// ObjectsHandler handles Objects requests
type ObjectsHandler struct {
	ObjectService    cabby.ObjectService
	StatusService    cabby.StatusService
	MaxContentLength int64
}

/* Get */

// Get handles a get request
func (h ObjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	objectID := takeObjectID(r)

	if objectID == "" {
		h.getObjects(w, r)
		return
	}
	h.getObject(w, r)
}

func (h ObjectsHandler) getObjects(w http.ResponseWriter, r *http.Request) {
	cr, err := cabby.NewRange(r.Header.Get("Range"))
	if err != nil {
		rangeNotSatisfiable(w, err)
		return
	}

	objects, err := h.ObjectService.Objects(takeCollectionID(r), &cr)
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

	if cr.Valid() {
		w.Header().Set("Content-Range", cr.String())
		writePartialContent(w, cabby.StixContentType, resourceToJSON(bundle))
	} else {
		writeContent(w, cabby.StixContentType, resourceToJSON(bundle))
	}
}

func (h ObjectsHandler) getObject(w http.ResponseWriter, r *http.Request) {
	object, err := h.ObjectService.Object(takeCollectionID(r), takeObjectID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if object.ID == "" {
		resourceNotFound(w, fmt.Errorf("Object ID doesn't exist in this collection"))
	} else {
		writeContent(w, cabby.TaxiiContentType, resourceToJSON(object))
	}
}

/* Post */

// Post handles post request
func (h ObjectsHandler) Post(w http.ResponseWriter, r *http.Request) {
	if greaterThan(r.ContentLength, h.MaxContentLength) {
		requestTooLarge(w, r.ContentLength, h.MaxContentLength)
		return
	}

	if !takeCollectionAccess(r).CanWrite {
		forbidden(w, fmt.Errorf("Unauthorized to write to collection"))
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

	err = h.StatusService.CreateStatus(status)
	if err != nil {
		internalServerError(w, errors.New("Unable to store status resource"))
		return
	}

	w.WriteHeader(http.StatusAccepted)
	writeContent(w, cabby.TaxiiContentType, resourceToJSON(status))
	go h.ObjectService.CreateBundle(bundle, takeCollectionID(r), status, h.StatusService)
}

/* helpers */

func bundleFromBytes(b []byte) (stones.Bundle, error) {
	var bundle stones.Bundle

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

func greaterThan(r, m int64) bool {
	if r > m {
		return true
	}
	return false
}

func objectsToBundle(objects []cabby.Object) (stones.Bundle, error) {
	bundle, err := stones.NewBundle()
	if err != nil {
		return bundle, err
	}

	for _, o := range objects {
		bundle.Objects = append(bundle.Objects, o.Object)
	}

	if len(bundle.Objects) == 0 {
		log.Warn("Can't return an empty bundle, returning error to caller")
		return bundle, errors.New("No data returned: empty bundle")
	}
	return bundle, err
}
