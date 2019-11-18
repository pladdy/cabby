package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// ManifestMethods lists allowed methods
const ManifestMethods = "Get, Head"

// ManifestHandler holds a cabby ManifestService
type ManifestHandler struct {
	ManifestService cabby.ManifestService
}

// Delete handler
func (h ManifestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, r, ManifestMethods)
}

// Get serves a manifest resource
func (h ManifestHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ManifestHandler"}).Debug("Handler called")

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

	manifest, err := h.ManifestService.Manifest(r.Context(), takeCollectionID(r), &p, newFilter(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if noResources(len(manifest.Objects)) {
		resourceNotFound(w, errors.New("No resources available for this request"))
		return
	}

	w.Header().Set("X-TAXII-Date-Added-First", p.AddedAfterFirst())
	w.Header().Set("X-TAXII-Date-Added-Last", p.AddedAfterLast())
	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(manifest))
}

// Post handler
func (h ManifestHandler) Post(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, r, ManifestMethods)
}
