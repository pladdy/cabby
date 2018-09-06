package cabby

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestLogServiceEnd(t *testing.T) {
	// redirect log output for test
	var buf bytes.Buffer

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(&buf)

	defer func() {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stderr)
	}()

	resource := t.Name()
	action := "test"
	LogServiceEnd(context.Background(), resource, action, time.Now().In(time.UTC))

	type expectedLog struct {
		Action    string
		ElapsedTS float64 `json:"elapsed_ts"`
		EndTS     int64   `json:"end_ts"`
		Level     string
		Msg       string
		Resource  string
		Time      string
	}

	// parse log into struct
	var result expectedLog
	err := json.Unmarshal([]byte(buf.String()), &result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Action != action {
		t.Error("Got:", result.Action, "Expected:", action)
	}
	if result.ElapsedTS <= 0 {
		t.Error("Got:", result.ElapsedTS, "Expected: > 0")
	}
	if result.EndTS <= 0 {
		t.Error("Got:", result.EndTS, "Expected: > 0")
	}
	if result.Level != "info" {
		t.Error("Got:", result.Level, "Expected:", "info")
	}
	if result.Msg != "Finished serving resource" {
		t.Error("Got:", result.Level, "Expected:", "Finished serving resource")
	}
	if result.Resource != resource {
		t.Error("Got:", result.Resource, "Expected:", resource)
	}
	if len(result.Time) <= 0 {
		t.Error("Got:", result.Time, "Expected: len > 0")
	}
}

func TestLogServiceStart(t *testing.T) {
	// redirect log output for test
	var buf bytes.Buffer

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(&buf)

	defer func() {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stderr)
	}()

	resource := t.Name()
	action := "test"
	_ = LogServiceStart(context.Background(), resource, action)

	type expectedLog struct {
		Action   string
		Level    string
		Msg      string
		Resource string
		StartTS  int64 `json:"start_ts"`
		Time     string
	}

	// parse log into struct
	var result expectedLog
	err := json.Unmarshal([]byte(buf.String()), &result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Action != action {
		t.Error("Got:", result.Action, "Expected:", action)
	}
	if result.Level != "info" {
		t.Error("Got:", result.Level, "Expected:", "info")
	}
	if result.Msg != "Serving resource" {
		t.Error("Got:", result.Level, "Expected:", "Serving resource")
	}
	if result.Resource != resource {
		t.Error("Got:", result.Resource, "Expected:", resource)
	}
	if result.StartTS <= 0 {
		t.Error("Got:", result.StartTS, "Expected: > 0")
	}
	if len(result.Time) <= 0 {
		t.Error("Got:", result.Time, "Expected: len > 0")
	}
}
