package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// APIRootHandler holds a cabby APIRootService
type APIRootHandler struct {
	APIRootService cabby.APIRootService
}

// Get handles a get request
func (h APIRootHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "APIRootHandler"}).Debug("Handler called")

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

// Post handles post request
func (h APIRootHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "Get, Head")
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
