package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
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
		return []cabby.APIRoot{tester.APIRoot}, errors.New("foo")
	}
	ds.APIRootServiceFn = func() tester.APIRootService { return as }

	registerAPIRoots(ds, sm)

	// parse log into struct
	var result requestLog
	err := json.Unmarshal([]byte(lastLog(buf)), &result)
	if err != nil {
		t.Fatal(err)
	}

	testErrorLog(result, t)
}
