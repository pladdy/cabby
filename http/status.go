package http

import (
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
)

// StatusMethods lists allowed methods
const StatusMethods = "Get, Head"

// StatusHandler holds a cabby StatusService
type StatusHandler struct {
	StatusService cabby.StatusService
}

// Delete handler
func (h StatusHandler) Delete(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, r, StatusMethods)
}

// Get serves a status resource
func (h StatusHandler) Get(w http.ResponseWriter, r *http.Request) {
	status, err := h.StatusService.Status(r.Context(), takeStatusID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if status.TotalCount <= 0 {
		resourceNotFound(w, errors.New("No status available for this id"))
		return
	}

	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(status))
}

// Post handler
func (h StatusHandler) Post(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, r, StatusMethods)
}
