package cabby

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func logResourceStart(resource, action string) time.Time {
	milliSecondOfNanoSeconds := int64(1000000)

	start := time.Now().In(time.UTC)
	log.WithFields(log.Fields{
		"action":   action,
		"resource": resource,
		"start_ts": start.UnixNano() / milliSecondOfNanoSeconds,
	}).Info("Starting with resource")
	return start
}

func logResourceEnd(resource, action string, start time.Time) {
	milliSecondOfNanoSeconds := int64(1000000)

	end := time.Now().In(time.UTC)
	elapsed := time.Since(start)

	log.WithFields(log.Fields{
		"action":     action,
		"elapsed_ts": float64(elapsed.Nanoseconds()) / float64(milliSecondOfNanoSeconds),
		"end_ts":     end.UnixNano() / milliSecondOfNanoSeconds,
		"resource":   resource,
	}).Info("Finished with resource")
}

// ReadFunc defines signature for a discovery reader
type ReadFunc func() (Result, error)

// WithReadLogging takes a resource name and a read function and adds logging to it
func WithReadLogging(resource string, f ReadFunc) ReadFunc {
	return func() (Result, error) {
		action := "read"
		start := logResourceStart(resource, action)
		result, err := f()
		logResourceEnd(resource, action, start)
		return result, err
	}
}
