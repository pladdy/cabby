package http

import (
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// VersionsMethods lists allowed methods
const VersionsMethods = "Get, Head"

// VersionsHandler holds a cabby VersionsService
type VersionsHandler struct {
	VersionsService cabby.VersionsService
}

// Delete handler
func (h VersionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, r, VersionsMethods)
}

// Get serves a Versions resource
func (h VersionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "VersionsHandler"}).Debug("Handler called")

	if !requestIsReadAuthorized(r) {
		forbidden(w, errors.New("Unauthorized access"))
		return
	}

	p, err := cabby.NewPage(takeLimit(r))
	if err != nil {
		badRequest(w, err)
		return
	}

	versions, err := h.VersionsService.Versions(r.Context(), takeCollectionID(r), takeObjectID(r), &p, newFilter(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if noResources(len(versions.Versions)) {
		resourceNotFound(w, errors.New("No resources available for this request"))
		return
	}

	w.Header().Set("X-TAXII-Date-Added-First", p.AddedAfterFirst())
	w.Header().Set("X-TAXII-Date-Added-Last", p.AddedAfterLast())
	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(versions))
}

// Post handler
func (h VersionsHandler) Post(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, r, VersionsMethods)
}
