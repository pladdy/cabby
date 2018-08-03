package http

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// func TestHeaders(t *testing.T) {
// 	ts := getStorer()
// 	defer ts.disconnect()
//
// 	h := handleTaxiiDiscovery(ts, testConfig().Port)
// 	s := httptest.NewServer(http.HandlerFunc(h))
// 	defer s.Close()
//
// 	res, err := http.Get(s.URL)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	result := res.Header["Content-Type"][0]
// 	expected := taxiiContentType
//
// 	if result != expected {
// 		t.Error("Got:", result, "Expected:", expected)
// 	}
// }

func TestRequireAcceptType(t *testing.T) {
	tests := []struct {
		acceptedHeader string
		acceptHeader   string
		responseCode   int
	}{
		{StixContentType, "application/vnd.oasis.stix+json; version=2.0", http.StatusOK},
		{StixContentType, "application/vnd.oasis.stix+json", http.StatusOK},
		{StixContentType, "application/vnd.oasis.stix+json;verion=2.0", http.StatusOK},
		{StixContentType, "", http.StatusUnsupportedMediaType},
		{StixContentType, "application/vnd.oasis.stix+jsonp", http.StatusUnsupportedMediaType},
		{StixContentType, "application/vnd.oasis.stix+jsonp; version=3.0", http.StatusUnsupportedMediaType},
		{TaxiiContentType, "application/vnd.oasis.taxii+json; version=2.0", http.StatusOK},
		{TaxiiContentType, "application/vnd.oasis.taxii+json", http.StatusOK},
		{TaxiiContentType, "application/vnd.oasis.taxii+json;verion=2.0", http.StatusOK},
		{TaxiiContentType, "", http.StatusUnsupportedMediaType},
		{TaxiiContentType, "application/vnd.oasis.taxii+jsonp", http.StatusUnsupportedMediaType},
		{TaxiiContentType, "application/vnd.oasis.taxii+jsonp; version=3.0", http.StatusUnsupportedMediaType},
	}

	for _, test := range tests {
		mockHandler := func(w http.ResponseWriter, r *http.Request) {
			accept := r.Header.Get("Accept")
			io.WriteString(w, fmt.Sprintf("Accept Header: %v", accept))
		}
		mockHandler = WithAcceptType(mockHandler, test.acceptedHeader)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Add("Accept", test.acceptHeader)
		res := httptest.NewRecorder()

		mockHandler(res, req)
		body, _ := ioutil.ReadAll(res.Body)

		if res.Code != test.responseCode {
			t.Error("Got:", res.Code, string(body), "Expected:", http.StatusOK)
		}
	}
}
