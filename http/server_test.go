package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
	log "github.com/sirupsen/logrus"
)

func TestNewCabby(t *testing.T) {
	c := cabby.Config{Port: 1212, SSLCert: "../server.crt", SSLKey: "../server.key"}
	server := NewCabby(mockDataStore(), c)
	defer server.Close()

	// use tls which requires cert/key files
	go func() {
		log.Info(server.ListenAndServeTLS(c.SSLCert, c.SSLKey))
	}()

	// send request
	req := newServerRequest("GET", "https://localhost:"+strconv.Itoa(c.Port)+"/taxii/")
	req.Header.Set("Accept", cabby.TaxiiContentType)

	client := tlsClient()
	res, err := attemptRequest(client, req)
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
	us.UserFn = func(ctx context.Context, user, password string) (cabby.User, error) {
		userCalled = true
		return cabby.User{Email: "foo"}, nil
	}

	us.UserCollectionsFn = func(ctx context.Context, user string) (cabby.UserCollectionList, error) {
		return cabby.UserCollectionList{}, nil
	}

	// set up a data store with mocked services
	ds := tester.DataStore{}
	ds.UserServiceFn = func() tester.UserService { return us }

	// create and register a handler on a test route
	h := testHandler(t.Name())
	sm := http.NewServeMux()
	sm.HandleFunc("/test/", h)

	port := 1212
	server := setupServer(ds, sm, cabby.Config{Port: port})
	defer server.Close()

	// ignore TLS, not needed for log test
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
	var result tester.RequestLog
	err = json.Unmarshal([]byte(tester.LastLog(buf)), &result)
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

	handler := http.NewServeMux()
	server := setupServer(mockDataStore(), handler, cabby.Config{Port: 1234})
	defer server.Close()

	type expectedLog struct {
		Time  string
		Level string
		Msg   string
		Port  string
	}

	// parse log into struct
	var result expectedLog
	err := json.Unmarshal([]byte(tester.LastLog(buf)), &result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Port != "1234" {
		t.Error("Got:", result.Port, "Expected:", "1234", "Log:", result)
	}
}

func TestSetupServerSettings(t *testing.T) {
	handler := http.NewServeMux()
	server := setupServer(mockDataStore(), handler, cabby.Config{Port: 1234})
	defer server.Close()

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
		// TLS 1.2
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: true,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   true,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: true,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    true,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  true,
		// TLS 1.3
		tls.TLS_AES_256_GCM_SHA384:                true,
		tls.TLS_CHACHA20_POLY1305_SHA256:          true,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256: true,
	}
	for _, cipherSuite := range tlsSetup.CipherSuites {
		if !expectedCipherSuites[cipherSuite] {
			t.Error("Invalid CipherSuite:", cipherSuite)
		}
	}
}
