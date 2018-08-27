package http

import (
	"errors"
	"net/http"

	cabby "github.com/pladdy/cabby2"
)

// StatusHandler holds a cabby StatusService
type StatusHandler struct {
	StatusService cabby.StatusService
}

// Get serves a status resource
func (h StatusHandler) Get(w http.ResponseWriter, r *http.Request) {
	status, err := h.StatusService.Status(takeStatusID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if status.TotalCount <= 0 {
		resourceNotFound(w, errors.New("No status available for this id"))
		return
	}

	writeContent(w, cabby.TaxiiContentType, resourceToJSON(status))
}

// Post handles post request
func (h StatusHandler) Post(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
