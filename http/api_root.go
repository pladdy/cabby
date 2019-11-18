package http

import (
	"fmt"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// APIRootMethods lists allowed methods
const APIRootMethods = "Get, Head"

// APIRootHandler holds a cabby APIRootService
type APIRootHandler struct {
	APIRootService cabby.APIRootService
}

// Delete handler
func (h APIRootHandler) Delete(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, r, APIRootMethods)
}

// Get handles a get request
func (h APIRootHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "APIRootHandler"}).Debug("Handler called")

	if !verifyRequestHeader(r, "Accept", cabby.TaxiiContentType) {
		notAcceptable(w, fmt.Errorf("Accept header must be '%v'", cabby.TaxiiContentType))
		return
	}

	path := trimSlashes(r.URL.Path)

	apiRoot, err := h.APIRootService.APIRoot(r.Context(), path)
	if err != nil {
		internalServerError(w, err)
		return
	}

	if apiRoot.Title == "" {
		resourceNotFound(w, fmt.Errorf("API Root not found"))
		return
	}

	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(apiRoot))
}

// Post handler
func (h APIRootHandler) Post(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, r, APIRootMethods)
}
