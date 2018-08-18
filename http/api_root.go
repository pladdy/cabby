package http

import (
	"fmt"
	"net/http"

	cabby "github.com/pladdy/cabby2"
)

// APIRootHandler holds a cabby APIRootService
type APIRootHandler struct {
	APIRootService cabby.APIRootService
}

// Get handles a get request
func (h APIRootHandler) Get(w http.ResponseWriter, r *http.Request) {
	path := trimSlashes(r.URL.Path)

	apiRoot, err := h.APIRootService.APIRoot(path)
	if err != nil {
		internalServerError(w, err)
		return
	}

	if apiRoot.Title == "" {
		resourceNotFound(w, fmt.Errorf("API Root not found"))
	} else {
		writeContent(w, cabby.TaxiiContentType, resourceToJSON(apiRoot))
	}
}
