package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
	log "github.com/sirupsen/logrus"
)

func TestNewCabby(t *testing.T) {
	// mock up services
	us := tester.UserService{}
	us.UserFn = func(user, password string) (cabby.User, error) {
		return cabby.User{}, nil
	}
	us.ExistsFn = func(cabby.User) bool { return true }

	as := tester.APIRootService{}
	as.APIRootsFn = func() ([]cabby.APIRoot, error) { return []cabby.APIRoot{cabby.APIRoot{}}, nil }

	cs := tester.CollectionService{}
	cs.CollectionFn = func(user, collectionID, apiRootPath string) (cabby.Collection, error) {
		return cabby.Collection{}, nil
	}
	cs.CollectionsFn = func(user, apiRootPath string) (cabby.Collections, error) { return cabby.Collections{}, nil }
	cs.CollectionsInAPIRootFn = func(apiRootPath string) (cabby.CollectionsInAPIRoot, error) {
		return cabby.CollectionsInAPIRoot{}, nil
	}

	ds := tester.DiscoveryService{}
	ds.DiscoveryFn = func() (cabby.Discovery, error) { return cabby.Discovery{Title: t.Name()}, nil }

	// set up a data store with mocked services
	md := tester.DataStore{}
	md.APIRootServiceFn = func() tester.APIRootService { return as }
	md.CollectionServiceFn = func() tester.CollectionService { return cs }
	md.DiscoveryServiceFn = func() tester.DiscoveryService { return ds }
	md.UserServiceFn = func() tester.UserService { return us }

	port := 78122
	server := NewCabby(md, cabby.Config{Port: port})
	defer server.Close()

	go func() {
		log.Info(server.ListenAndServe())
	}()

	// send request
	client := http.Client{}
	req := newServerRequest("GET", "http://localhost:"+strconv.Itoa(port)+"/taxii/")
	req.Header.Set("Accept", cabby.TaxiiContentType)

	res, err := attemptRequest(&client, req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Error("Got:", res.StatusCode, "Expected:", http.StatusOK)
	}
}

func TestSetupServerHandler(t *testing.T) {
	// redirect log output for test
	var buf bytes.Buffer

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(&buf)

	defer func() {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stderr)
	}()

	// mock up service; add a variable to track if User() is called
	us := tester.UserService{}
	userCalled := false

	us.UserFn = func(user, password string) (cabby.User, error) {
		userCalled = true
		return cabby.User{}, nil
	}
	us.ExistsFn = func(cabby.User) bool { return true }

	// set up a data store with mocked services
	ds := tester.DataStore{}
	ds.UserServiceFn = func() tester.UserService { return us }

	// create and register a handler on a test route
	h := testHandler(t.Name())
	sm := http.NewServeMux()
	sm.HandleFunc("/test/", h)

	port := 78122
	server := setupServer(ds, sm, cabby.Config{Port: port})
	defer server.Close()

	go func() {
		log.Info(server.ListenAndServe())
	}()

	// send request
	client := http.Client{}
	req := newServerRequest("GET", "http://localhost:"+strconv.Itoa(port)+"/test/")
	res, err := attemptRequest(&client, req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Error("Got:", res.StatusCode, "Expected:", http.StatusOK)
	}

	// parse log into struct
	var result requestLog
	err = json.Unmarshal([]byte(lastLog(buf)), &result)
	if err != nil {
		t.Fatal(err)
	}

	testRequestLog(result, t)

	// server handlers require basic auth
	if userCalled != true {
		t.Error("Got:", userCalled, "Expected: true")
	}
}

func TestSetupServerLogging(t *testing.T) {
	// redirect log output for test
	var buf bytes.Buffer

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(&buf)

	defer func() {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stderr)
	}()

	// set up test
	ds := tester.DataStore{}
	ds.UserServiceFn = func() tester.UserService {
		return tester.UserService{}
	}

	handler := http.NewServeMux()
	_ = setupServer(ds, handler, cabby.Config{Port: 1234})

	type expectedLog struct {
		Time  string
		Level string
		Msg   string
		Port  string
	}

	// parse log into struct
	var result expectedLog
	err := json.Unmarshal([]byte(lastLog(buf)), &result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Port != "1234" {
		t.Error("Got:", result.Port, "Expected:", "1234")
	}
}

func TestSetupServerSettings(t *testing.T) {
	// set up test
	ds := tester.DataStore{}
	ds.UserServiceFn = func() tester.UserService {
		return tester.UserService{}
	}

	handler := http.NewServeMux()
	server := setupServer(ds, handler, cabby.Config{Port: 1234})

	// set server settings
	expectedAddr := ":1234"
	if server.Addr != expectedAddr {
		t.Error("Got:", server.Addr, "Expected:", expectedAddr)
	}
	if server.TLSConfig == nil {
		t.Error("TLSConfig should not be nil")
	}
	if server.TLSNextProto == nil {
		t.Error("TLSNextProto should not be nil")
	}
}

func TestSetupTLS(t *testing.T) {
	tlsSetup := setupTLS()

	if tlsSetup.MinVersion != tls.VersionTLS12 {
		t.Error("Got:", tlsSetup.MinVersion, "Expected:", tls.VersionTLS12)
	}

	expectedCurves := map[tls.CurveID]bool{
		tls.CurveP521: true,
		tls.CurveP384: true,
		tls.CurveP256: true,
	}
	for _, curve := range tlsSetup.CurvePreferences {
		if !expectedCurves[curve] {
			t.Error("Invalid CurvePreference:", curve)
		}
	}

	if tlsSetup.PreferServerCipherSuites != true {
		t.Error("Got:", tlsSetup.PreferServerCipherSuites, "Expected:", true)
	}

	expectedCipherSuites := map[uint16]bool{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384: true,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:    true,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384:       true,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA:          true,
	}
	for _, cipherSuite := range tlsSetup.CipherSuites {
		if !expectedCipherSuites[cipherSuite] {
			t.Error("Invalid CurvePreference:", cipherSuite)
		}
	}
}
