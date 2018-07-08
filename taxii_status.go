package main

import (
	"errors"
	"net/http"
)

/* handler */

func handleTaxiiStatus(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusID, err := taxiiIDFromString(takeStatusID(r))
		if err != nil {
			resourceNotFound(w, err)
			return
		}

		status := taxiiStatus{ID: statusID}
		err = status.read(ts)
		if err != nil {
			resourceNotFound(w, err)
			return
		}

		writeContent(w, taxiiContentType, resourceToJSON(status))
	})
}

/* model */

type taxiiStatus struct {
	ID               taxiiID  `json:"id"`
	Status           string   `json:"status"`
	RequestTimestamp string   `json:"request_timestamp"`
	TotalCount       int64    `json:"total_count"`
	SuccessCount     int64    `json:"success_count"`
	Successes        []string `json:"successes"`
	FailureCount     int64    `json:"failure_count"`
	Failures         []string `json:"failures"`
	PendingCount     int64    `json:"pending_count"`
	Pendings         []string `json:"pendings"`
}

func newTaxiiStatus(objects int) (taxiiStatus, error) {
	if objects < 1 {
		return taxiiStatus{}, errors.New("Can't post less than 1 object")
	}

	id, err := newTaxiiID()
	if err != nil {
		return taxiiStatus{}, err
	}

	count := int64(objects)
	return taxiiStatus{ID: id, Status: "pending", TotalCount: count, PendingCount: count}, err
}

func (s *taxiiStatus) create(ts taxiiStorer) error {
	err := createResource(
		ts,
		"taxiiStatus",
		[]interface{}{s.ID, s.Status, s.TotalCount, s.SuccessCount, s.FailureCount, s.PendingCount})

	return err
}

func (s *taxiiStatus) read(ts taxiiStorer) error {
	status := *s

	result, err := ts.read("taxiiStatus", []interface{}{s.ID.String()})
	if err != nil {
		return err
	}
	status = result.data.(taxiiStatus)

	*s = status
	return err
}

func (s *taxiiStatus) update(ts taxiiStorer, status string) error {
	err := ts.update(
		"taxiiStatus",
		[]interface{}{status, s.TotalCount, s.SuccessCount, s.FailureCount, s.PendingCount, s.ID})

	return err
}
