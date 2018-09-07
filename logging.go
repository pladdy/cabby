package cabby

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

const milliSecondOfNanoSeconds = int64(1000000)

// LogServiceStart takes a resource and action being performed and logs it
func LogServiceStart(ctx context.Context, resource, action string) time.Time {
	start := time.Now().In(time.UTC)
	log.WithFields(log.Fields{
		"action":         action,
		"resource":       resource,
		"start_ts":       start.UnixNano() / milliSecondOfNanoSeconds,
		"transaction_id": TakeTransactionID(ctx).String(),
		"user":           TakeUser(ctx).Email,
	}).Info("Serving resource")
	return start
}

// LogServiceEnd takes a resource and action being performed and a start time, and logs it and how long it took
func LogServiceEnd(ctx context.Context, resource, action string, start time.Time) {
	end := time.Now().In(time.UTC)
	elapsed := time.Since(start)

	log.WithFields(log.Fields{
		"action":         action,
		"elapsed_ts":     float64(elapsed.Nanoseconds()) / float64(milliSecondOfNanoSeconds),
		"end_ts":         end.UnixNano() / milliSecondOfNanoSeconds,
		"resource":       resource,
		"transaction_id": TakeTransactionID(ctx).String(),
		"user":           TakeUser(ctx).Email,
	}).Info("Finished serving resource")
}
