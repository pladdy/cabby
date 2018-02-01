package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func init() {
	reloadTestConfig()
}

/* helpers */

func attemptRequest(c *http.Client, r *http.Request) (*http.Response, error) {
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
	renameFile(configPath, configPath+".testing")
	renameFile("test/config/testing_config.json", configPath)

	defer func() {
		renameFile(configPath, "test/config/testing_config.json")
		renameFile(configPath+".testing", configPath)
	}()

	server := newCabby()
	go func() {
		server.ListenAndServeTLS(config.SSLCert, config.SSLKey)
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

func TestHSTS(t *testing.T) {
	req := requestWithBasicAuth(discoveryURL)
	res, _ := requestFromTestServer(req)

	expected := "max-age=" + sixMonthsOfSeconds + "; includeSubDomains"
	result := strings.Join(res.Header["Strict-Transport-Security"], "")

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestBasicAuth(t *testing.T) {
	// set up the expected discovery
	cd := config.Discovery

	cd.Default = insertPort(cd.Default)

	// add port to api roots
	var apiRootsWithPort []string
	for _, apiRoot := range cd.APIRoots {
		apiRootsWithPort = append(apiRootsWithPort, insertPort(apiRoot))
	}
	cd.APIRoots = apiRootsWithPort

	expectedDiscovery, err := json.Marshal(cd)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		user           string
		pass           string
		expectedStatus int
		expectedBody   string
	}{
		{testUser, testPass, 200, string(expectedDiscovery)},
		{"invalid", "pass", 401, `{"title":"Unauthorized","description":"Invalid user/pass combination","http_status":"401"}` + "\n"},
	}

	for _, test := range tests {
		req, err := http.NewRequest("GET", discoveryURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(test.user, test.pass)

		res, body := requestFromTestServer(req)

		if res.StatusCode != test.expectedStatus {
			t.Error("Got:", res.StatusCode, "Expected:", test.expectedStatus)
		}
		if body != test.expectedBody {
			t.Error("Got:", body, "Expected:", test.expectedBody)
		}
	}
}

func TestMain(t *testing.T) {
	// use test config
	renameFile(configPath, configPath+".testing")
	renameFile("test/config/different_port_config.json", configPath)

	defer func() {
		renameFile(configPath, "test/config/different_port_config.json")
		renameFile(configPath+".testing", configPath)
	}()

	go func() {
		main()
	}()

	req := requestWithBasicAuth(discoveryURL)
	res, _ := attemptRequest(tlsClient(), req)
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	var result taxiiDiscovery
	err := json.Unmarshal(body, &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := config.Discovery
	expected.Default = insertPort(expected.Default)

	if result.Default != expected.Default {
		t.Error("Got:", result.Default, "Expected:", expected.Default)
	}
}

func TestMainAPIRoot(t *testing.T) {
	req := requestWithBasicAuth(apiRootURL)
	_, body := requestFromTestServer(req)

	u, _ := url.Parse(apiRootURL)
	noPortHost := urlWithNoPort(u)

	config := cabbyConfig{}.parse(configPath)
	expected, _ := json.Marshal(config.APIRootMap[noPortHost])

	if body != string(expected) {
		t.Error("Got:", body, "Expected:", string(expected))
	}
}

func TestRegisterAPIRootInvalidURL(t *testing.T) {
	recovered := false

	defer func() {
		if err := recover(); err == nil {
			t.Error("Failed to recover")
		}
		recovered = true
	}()

	mockHandler := http.NewServeMux()
	registerAPIRoot("", mockHandler)

	if recovered != true {
		t.Error("Expected: 'recovered' to be true")
	}
}
