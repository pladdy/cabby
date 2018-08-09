package cabby

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const milliSecondOfNanoSeconds = int64(1000000)

// LogServiceStart takes a resource and action being performed and logs it
func LogServiceStart(resource, action string) time.Time {
	start := time.Now().In(time.UTC)
	log.WithFields(log.Fields{
		"action":   action,
		"resource": resource,
		"start_ts": start.UnixNano() / milliSecondOfNanoSeconds,
	}).Info("Serving resource")
	return start
}

// LogServiceEnd takes a resource and action being performed and a start time, and logs it and how long it took
func LogServiceEnd(resource, action string, start time.Time) {
	end := time.Now().In(time.UTC)
	elapsed := time.Since(start)

	log.WithFields(log.Fields{
		"action":     action,
		"elapsed_ts": float64(elapsed.Nanoseconds()) / float64(milliSecondOfNanoSeconds),
		"end_ts":     end.UnixNano() / milliSecondOfNanoSeconds,
		"resource":   resource,
	}).Info("Finished serving resource")
}
