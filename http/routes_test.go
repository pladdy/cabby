package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
	log "github.com/sirupsen/logrus"
)

func TestRegisterAPIRoutesFail(t *testing.T) {
	// redirect log output for test
	var buf bytes.Buffer

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(&buf)

	defer func() {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stderr)
	}()

	// create a server
	sm := http.NewServeMux()

	// mock out services and have api roots fail
	ds := mockDataStore()
	as := tester.APIRootService{}
	as.APIRootsFn = func(ctx context.Context) ([]cabby.APIRoot, error) {
		return []cabby.APIRoot{tester.APIRoot}, errors.New("service error")
	}
	ds.APIRootServiceFn = func() tester.APIRootService { return as }

	registerAPIRoots(ds, sm)

	// parse log into struct
	var result tester.RequestLog
	err := json.Unmarshal([]byte(tester.LastLog(buf)), &result)
	if err != nil {
		t.Fatal(err)
	}

	testErrorLog(result, t)
}

func TestRegisterCollectionRoutesFail(t *testing.T) {
	// redirect log output for test
	var buf bytes.Buffer

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(&buf)

	defer func() {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stderr)
	}()

	// create a server
	sm := http.NewServeMux()

	// mock out services and have api roots fail
	ds := mockDataStore()
	cs := tester.CollectionService{}
	cs.CollectionsInAPIRootFn = func(ctx context.Context, path string) (cabby.CollectionsInAPIRoot, error) {
		return cabby.CollectionsInAPIRoot{}, errors.New("service error")
	}
	ds.CollectionServiceFn = func() tester.CollectionService { return cs }

	registerCollectionRoutes(ds, cabby.APIRoot{}, sm)

	// parse log into struct
	var result tester.RequestLog
	err := json.Unmarshal([]byte(tester.LastLog(buf)), &result)
	if err != nil {
		t.Fatal(err)
	}

	testErrorLog(result, t)
}

func TestRouteRequest(t *testing.T) {
	mock := mockRequestHandler{}

	tests := []struct {
		method string
		url    string
		status int
	}{
		{"CUSTOM", tester.BaseURL, http.StatusMethodNotAllowed},
		{http.MethodGet, tester.BaseURL, http.StatusOK},
		{http.MethodHead, tester.BaseURL, http.StatusOK},
		{http.MethodPost, tester.BaseURL, http.StatusOK},
		{http.MethodDelete, tester.BaseURL, http.StatusOK},
	}

	for _, test := range tests {
		status, _ := handlerTest(routeRequest(mock), test.method, test.url, nil)

		if status != test.status {
			t.Error("Testing method: ", test.method, "Got:", status, "Expected:", test.status)
		}
	}
}
