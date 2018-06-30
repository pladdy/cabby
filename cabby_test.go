package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

/* helpers */

func attemptRequest(c *http.Client, r *http.Request) (*http.Response, error) {
	info.Println("Requesting", r.URL, "from test server")
	MaxTries := 3

	for i := 0; i < MaxTries; i++ {
		res, err := c.Do(r)
		if err != err {
			fail.Fatal(err)
		}
		if res != nil {
			return res, err
		}
		warn.Println("Web server for test not responding, waiting...")
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return nil, errors.New("Failed to get request")
}

func requestWithBasicAuth(u string) *http.Request {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		fail.Fatal(err)
	}
	req.Header.Add("Accept", taxiiContentType)
	req.SetBasicAuth(testUser, testPass)
	return req
}

func requestFromTestServer(r *http.Request) (*http.Response, string) {
	server := runTestServer()
	defer server.Close()

	res, err := attemptRequest(tlsClient(), r)
	if err != nil {
		fail.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != err {
		fail.Fatal(err)
	}

	return res, string(body)
}

// run server in go routine and return it
func runTestServer() *http.Server {
	server, err := newCabby(testConfig())
	if err != nil {
		fail.Fatal(err)
	}

	go func() {
		server.ListenAndServeTLS(testConfig().SSLCert, testConfig().SSLKey)
	}()
	return server
}

// set up a http client that uses TLS
func tlsClient() *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: tr}
}

/* tests */

func TestBasicAuth(t *testing.T) {
	tests := []struct {
		user           string
		pass           string
		expectedStatus int
	}{
		{testUser, testPass, http.StatusOK},
		{"invalid", "pass", http.StatusUnauthorized},
	}

	for _, test := range tests {
		req, err := http.NewRequest("GET", discoveryURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Accept", taxiiContentType)
		req.SetBasicAuth(test.user, test.pass)

		res, _ := requestFromTestServer(req)

		if res.StatusCode != test.expectedStatus {
			t.Error("Got:", res.StatusCode, "Expected:", test.expectedStatus)
		}
	}
}

func TestHSTS(t *testing.T) {
	req := requestWithBasicAuth(discoveryURL)
	res, _ := requestFromTestServer(req)

	expected := "max-age=" + sixMonthsOfSeconds + "; includeSubDomains"
	result := strings.Join(res.Header["Strict-Transport-Security"], "")

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestMain(t *testing.T) {
	renameFile(defaultDevelopmentConfig, defaultDevelopmentConfig+".testing")
	renameFile("testdata/config/main_test_config.json", defaultDevelopmentConfig)

	defer func() {
		renameFile(defaultDevelopmentConfig, "testdata/config/main_test_config.json")
		renameFile(defaultDevelopmentConfig+".testing", defaultDevelopmentConfig)
	}()

	go func() {
		main()
	}()

	mainTestConfigPort := "1235"

	req := requestWithBasicAuth(strings.Replace(discoveryURL, "1234", mainTestConfigPort, 1))
	res, _ := attemptRequest(tlsClient(), req)
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	var result taxiiDiscovery
	err := json.Unmarshal(body, &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := testDiscovery

	expected.Default = ""
	result.Default = ""

	if result.Default != expected.Default {
		t.Error("Got:", result.Default, "Expected:", expected.Default)
	}
}

func TestMainPanic(t *testing.T) {
	renameFile(defaultDevelopmentConfig, defaultDevelopmentConfig+".testing")
	renameFile("testdata/config/no_datastore_config.json", defaultDevelopmentConfig)

	defer func() {
		renameFile(defaultDevelopmentConfig, "testdata/config/no_datastore_config.json")
		renameFile(defaultDevelopmentConfig+".testing", defaultDevelopmentConfig)
	}()

	p := panicChecker{}
	go func() {
		defer attemptRecover(t, &p)

		main()
	}()

	// try to check the panic up to 3 times
	for i := 1; i <= 3; i++ {
		if p.recovered == false {
			time.Sleep(time.Duration(i*100) * time.Millisecond)
		}
	}
	if p.recovered != true {
		t.Error("Failed to recover a panic")
	}
}

func TestMainAPIRoot(t *testing.T) {
	setupSQLite()

	req := requestWithBasicAuth(testAPIRootURL)
	_, body := requestFromTestServer(req)

	var result taxiiAPIRoot
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := testAPIRoot

	if result.Title != expected.Title {
		t.Error("Got:", result.Title, "Expected:", expected.Title, "Result:", result)
	}
}

func TestNewCabbyNoAPIRoots(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_api_root")
	if err != nil {
		t.Fatal(err)
	}

	_, err = newCabby(testConfig())
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestNewCabbyFail(t *testing.T) {
	cs := configs{}.parse("testdata/config/no_datastore_config.json")
	c := cs["testing"]

	_, err := newCabby(c)
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestRegisterAPIRootInvalidPath(t *testing.T) {
	defer setupSQLite()

	// remove required table
	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_api_root")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	invalidPath := "foo"
	handler := http.NewServeMux()
	registerAPIRoot(ts, invalidPath, handler)
}
