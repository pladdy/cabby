package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestTakeAddedAfter(t *testing.T) {
	tests := []struct {
		request    *http.Request
		addedAfter string
	}{
		{httptest.NewRequest("GET", "/foo/bar/baz", nil), "0001-01-01T00:00:00Z"},
		{httptest.NewRequest("GET", "/foo/bar/baz?added_after=invalid", nil), "0001-01-01T00:00:00Z"},
		{httptest.NewRequest("GET", "/foo/bar/baz?added_after=2016-02-21T05:01:01.000Z", nil), "2016-02-21T05:01:01Z"},
		{httptest.NewRequest("GET", "/foo/bar/baz?added_after=2016-02-21T05:01:01.123Z", nil), "2016-02-21T05:01:01.123Z"},
	}

	for _, test := range tests {
		result := takeAddedAfter(test.request)
		if result.String() != test.addedAfter {
			t.Error("Got:", result.String(), "Expected:", test.addedAfter)
		}
	}
}

func TestTakeAPIRoot(t *testing.T) {
	tests := []struct {
		request  *http.Request
		expected string
	}{
		{httptest.NewRequest("GET", "/api_root/collections/collection_id/objects/stix_id/", nil), "api_root"},
		{httptest.NewRequest("GET", "/multi/token/api_root/collections/collection_id/objects/stix_id", nil), "multi/token/api_root"},
		{httptest.NewRequest("GET", "/invalid/foobar/collection_id/objects/stix_id", nil), ""},
	}

	for _, test := range tests {
		result := takeAPIRoot(test.request)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestTakeCollectionAccessInvalidCollection(t *testing.T) {
	// create a request with a valid context BUT a path with an invalid collection in it
	req := httptest.NewRequest("GET", "/foo/bar/baz", nil)

	ca := takeCollectionAccess(req)
	empty := cabby.CollectionAccess{}

	if ca != empty {
		t.Error("Got:", ca, "Expected:", empty)
	}
}

func TestTakeCollectionID(t *testing.T) {
	cid := tester.CollectionID

	tests := []struct {
		request *http.Request
		id      string
	}{
		{httptest.NewRequest("GET", "/api_root-path/collections/"+cid, nil), cid},
		{httptest.NewRequest("GET", "/api/root/path/collections/"+cid, nil), cid},
		{httptest.NewRequest("GET", "/api_root-path/collections/"+cid+"/objects", nil), cid},
		{httptest.NewRequest("GET", "/api/root/path/collections/"+cid+"/objects", nil), cid},
		{httptest.NewRequest("GET", "/api_root-path/collections/", nil), ""},
	}

	for _, test := range tests {
		result := takeCollectionID(test.request)
		if result != test.id {
			t.Error("Got:", result, "Expected:", test.id)
		}
	}
}

func TestTakeMatchIDs(t *testing.T) {
	tests := []struct {
		request *http.Request
		matchID string
	}{
		{httptest.NewRequest("GET", "/foo/bar/baz", nil), ""},
		{httptest.NewRequest("GET", "/foo/bar/baz?match[id]=some-id", nil), "some-id"},
		{httptest.NewRequest("GET", "/foo/bar/baz?match[id]=id1,id2,id3", nil), "id1,id2,id3"},
		{httptest.NewRequest("GET", "/foo/bar/baz?match[id]=id1&match[id]=id2&match[id]=id3", nil), "id1,id2,id3"},
	}

	for _, test := range tests {
		result := takeMatchIDs(test.request)
		if result != test.matchID {
			t.Error("Got:", result, "Expected:", test.matchID)
		}
	}
}

func TestTakeMatchTypes(t *testing.T) {
	tests := []struct {
		request   *http.Request
		matchType string
	}{
		{httptest.NewRequest("GET", "/foo/bar/baz", nil), ""},
		{httptest.NewRequest("GET", "/foo/bar/baz?match[type]=some-type", nil), "some-type"},
		{httptest.NewRequest("GET", "/foo/bar/baz?match[type]=type1,type2,type3", nil), "type1,type2,type3"},
		{httptest.NewRequest("GET", "/foo/bar/baz?match[type]=type1&match[type]=type2&match[type]=type3", nil),
			"type1,type2,type3"},
	}

	for _, test := range tests {
		result := takeMatchTypes(test.request)
		if result != test.matchType {
			t.Error("Got:", result, "Expected:", test.matchType)
		}
	}
}

func TestTakeMatchVersions(t *testing.T) {
	tests := []struct {
		request      *http.Request
		matchVersion string
	}{
		{httptest.NewRequest("GET", "/foo/bar/baz", nil), ""},
		{httptest.NewRequest("GET", "/foo/bar/baz?match[version]=some-version", nil), "some-version"},
		{httptest.NewRequest("GET", "/foo/bar/baz?match[version]=version1,version2,version3", nil),
			"version1,version2,version3"},
		{httptest.NewRequest("GET", "/foo/bar/baz?match[version]=version1&match[version]=version2&match[version]=version3", nil),
			"version1,version2,version3"},
	}

	for _, test := range tests {
		result := takeMatchVersions(test.request)
		if result != test.matchVersion {
			t.Error("Got:", result, "Expected:", test.matchVersion)
		}
	}
}

func TestTakeStatusID(t *testing.T) {
	sid := tester.StatusID

	tests := []struct {
		request *http.Request
		id      string
	}{
		{httptest.NewRequest("GET", "/api_root-path/status/"+sid, nil), sid},
		{httptest.NewRequest("GET", "/api/root/path/status/"+sid, nil), sid},
		{httptest.NewRequest("GET", "/api_root-path/status/"+sid+"/", nil), sid},
		{httptest.NewRequest("GET", "/api/root/path/status/"+sid+"/", nil), sid},
		{httptest.NewRequest("GET", "/api_root-path/collections/", nil), ""},
	}

	for _, test := range tests {
		result := takeStatusID(test.request)
		if result != test.id {
			t.Error("Got:", result, "Expected:", test.id)
		}
	}
}

func TestTakeObjectID(t *testing.T) {
	cid := tester.CollectionID
	oid := tester.ObjectID

	tests := []struct {
		request *http.Request
		result  string
	}{
		{httptest.NewRequest("GET", "/api_root-path/collections/"+cid+"/objects/"+oid+"/", nil), oid},
		{httptest.NewRequest("GET", "/api/root/path/collections/"+cid+"/objects/"+oid+"/", nil), oid},
		{httptest.NewRequest("GET", "/api_root-path/collections/"+cid+"/objects/"+oid, nil), oid},
		{httptest.NewRequest("GET", "/api/root/path/collections/"+cid+"/objects/"+oid, nil), oid},
		{httptest.NewRequest("GET", "/api_root-path/collections/", nil), ""},
	}

	for _, test := range tests {
		result := takeObjectID(test.request)
		if result != test.result {
			t.Error("Got:", result, "Expected:", test.result)
		}
	}
}

func TestTakeVersions(t *testing.T) {
	cid := tester.CollectionID
	oid := tester.ObjectID

	tests := []struct {
		request *http.Request
		result  string
	}{
		{httptest.NewRequest("GET", "/api_root-path/collections/"+cid+"/objects/"+oid+"/versions", nil), "versions"},
		{httptest.NewRequest("GET", "/api/root/path/collections/"+cid+"/objects/"+oid+"/versions", nil), "versions"},
		{httptest.NewRequest("GET", "/api_root-path/collections/"+cid+"/objects/"+oid+"/versions/", nil), "versions"},
		{httptest.NewRequest("GET", "/api/root/path/collections/"+cid+"/objects/"+oid+"/versions/", nil), "versions"},
		{httptest.NewRequest("GET", "/api_root-path/collections/", nil), ""},
	}

	for _, test := range tests {
		result := takeVersions(test.request)
		if result != test.result {
			t.Error("Got:", result, "Expected:", test.result)
		}
	}
}
