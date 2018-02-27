package main

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

func newTaxiiStatus() (taxiiStatus, error) {
	id, err := newTaxiiID()
	if err != nil {
		return taxiiStatus{}, err
	}

	// TODO: need to persist the id to a status table...
	return taxiiStatus{ID: id}, err
}
