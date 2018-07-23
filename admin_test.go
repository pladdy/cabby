// a lot of these tests are duplicative, but i'm doing that in the hope that it's more clear
// what's being tested.  that might be a dumb idea and instead there should be DRY'er tests below
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

var methodsToTest = []string{"DELETE", "POST", "PUT"}

var contextTests = []struct {
	contexts map[key]string
	status   int
}{
	{map[key]string{userName: testUser}, http.StatusForbidden},
	{map[key]string{}, http.StatusUnauthorized},
}

/* api root tests */

func TestHandleAdminTaxiiAPIRootInvalidMethod(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	handler := http.NewServeMux()
	status, _ := handlerTest(handleAdminTaxiiAPIRoot(ts, handler), "CUSTOM", testAdminAPIRootURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestHandleAdminTaxiiAPIRoot(t *testing.T) {
	setupSQLite()

	tests := []struct {
		method   string
		data     taxiiAPIRoot
		expected taxiiAPIRoot
	}{
		{method: "DELETE", data: taxiiAPIRoot{Path: t.Name(), Title: "deleted"}, expected: taxiiAPIRoot{Path: "", Title: ""}},
		{method: "POST", data: taxiiAPIRoot{Path: t.Name(), Title: "posted"}, expected: taxiiAPIRoot{Path: t.Name(), Title: "posted"}},
		{method: "PUT", data: taxiiAPIRoot{Path: t.Name(), Title: "updated"}, expected: taxiiAPIRoot{Path: t.Name(), Title: "updated"}},
	}

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		handler := http.NewServeMux()

		payload, err := json.Marshal(test.data)
		if err != nil {
			t.Fatal(err)
		}

		b := bytes.NewBuffer([]byte(payload))
		status, _ := handlerTest(handleAdminTaxiiAPIRoot(ts, handler), test.method, testAdminAPIRootURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK)
		}

		testTar := test.data
		result := taxiiAPIRoot{Path: testTar.Path}

		err = result.read(ts)
		if err != nil {
			t.Fatal(err)
		}

		expected := test.expected

		if result.Path != expected.Path {
			t.Error("Path:", result.Path, "Expected:", expected.Path, "Method:", test.method)
		}
		if result.Title != expected.Title {
			t.Error("Title:", result.Title, "Expected:", expected.Title, "Method:", test.method)
		}
	}
}

func TestAttemptRegisterAPIRoot(t *testing.T) {
	// post, delete, post and expect logging due to shared handler getting routes it already has
	setupSQLite()
	handler := http.NewServeMux()

	var buf bytes.Buffer

	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	tests := []struct {
		method  string
		payload string
	}{
		{method: "POST", payload: `{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`},
		{method: "DELETE", payload: `{"path": "` + t.Name() + `", "title": "` + "updated" + `"}`},
		{method: "POST", payload: `{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`},
	}

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		b := bytes.NewBuffer([]byte(test.payload))
		status, _ := handlerTest(handleAdminTaxiiAPIRoot(ts, handler), test.method, testAdminAPIRootURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK, "Method:", test.method)
		}
	}

	// check for warning in logs
	logs := regexp.MustCompile("\n").Split(strings.TrimSpace(buf.String()), -1)
	lastLog := logs[len(logs)-1]

	if match, _ := regexp.Match("failed to register api root handlers", []byte(lastLog)); !match {
		t.Error("Expected log output")
	}
}

func TestHandleAdminTaxiiAPIRootFailAuth(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	handler := http.NewServeMux()
	b := bytes.NewBuffer([]byte(`{"this won't get processed": true}`))

	for _, method := range methodsToTest {
		for _, test := range contextTests {
			ctx := testingContext()

			for k, v := range test.contexts {
				ctx = context.WithValue(ctx, k, v)
			}

			req := httptest.NewRequest(method, testAdminAPIRootURL, b)
			req = req.WithContext(ctx)
			res := httptest.NewRecorder()
			handleAdminTaxiiAPIRoot(ts, handler)(res, req)

			if res.Code != test.status {
				t.Error("Got:", res.Code, "Expected:", test.status)
			}
		}
	}
}

func TestHandleAdminTaxiiAPIRootFailData(t *testing.T) {
	setupSQLite()

	for _, method := range methodsToTest {
		ts := getStorer()
		defer ts.disconnect()

		handler := http.NewServeMux()

		b := bytes.NewBuffer([]byte(`{"pa"` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiAPIRoot(ts, handler), method, testAdminAPIRootURL, b)

		if status != http.StatusBadRequest {
			t.Error("Got:", status, "Expected:", http.StatusBadRequest)
		}
	}
}

func TestHandleAdminTaxiiAPIRootFailServer(t *testing.T) {
	tearDownSQLite()
	ts := getStorer()
	defer ts.disconnect()

	for _, method := range methodsToTest {
		handler := http.NewServeMux()

		b := bytes.NewBuffer([]byte(`{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiAPIRoot(ts, handler), method, testAdminAPIRootURL, b)

		if status != http.StatusInternalServerError {
			t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
		}
	}
}

/* collections */

func TestHandleAdminTaxiiCollectionsInvalidMethod(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleAdminTaxiiCollections(ts), "CUSTOM", testAdminCollectionsURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestHandleAdminTaxiiCollections(t *testing.T) {
	setupSQLite()

	id, err := taxiiIDFromString("cd9552e7-fb5d-4628-a724-0772ed51200c")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		method   string
		data     taxiiCollection
		expected taxiiCollection
	}{
		{method: "DELETE", data: taxiiCollection{ID: id, Title: "deleted"}, expected: taxiiCollection{}},
		{method: "POST", data: taxiiCollection{ID: id, Title: "posted"}, expected: taxiiCollection{ID: id, Title: "posted"}},
		{method: "PUT", data: taxiiCollection{ID: id, Title: "updated"}, expected: taxiiCollection{ID: id, Title: "updated"}},
	}

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		payload, err := json.Marshal(test.data)
		if err != nil {
			t.Fatal(err)
		}

		b := bytes.NewBuffer([]byte(payload))
		status, _ := handlerTest(handleAdminTaxiiCollections(ts), test.method, testAdminCollectionsURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK)
		}

		testTC := test.data
		result := taxiiCollection{ID: testTC.ID}

		_, err = result.read(ts, testUser)
		if err != nil {
			t.Fatal(err)
		}

		expected := test.expected

		if result.ID != expected.ID {
			t.Error("Path:", result.ID, "Expected:", expected.ID, "Method:", test.method)
		}
		if result.Title != expected.Title {
			t.Error("Title:", result.Title, "Expected:", expected.Title, "Method:", test.method)
		}
	}
}

func TestHandleAdminTaxiiCollectionsFailAuth(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	b := bytes.NewBuffer([]byte(`{"this won't get processed": true}`))

	for _, method := range methodsToTest {
		for _, test := range contextTests {
			ctx := testingContext()

			for k, v := range test.contexts {
				ctx = context.WithValue(ctx, k, v)
			}

			req := httptest.NewRequest(method, testAdminCollectionsURL, b)
			req = req.WithContext(ctx)
			res := httptest.NewRecorder()
			handleAdminTaxiiCollections(ts)(res, req)

			if res.Code != test.status {
				t.Error("Got:", res.Code, "Expected:", test.status)
			}
		}
	}
}

func TestHandleAdminTaxiiCollectionsFailData(t *testing.T) {
	setupSQLite()

	for _, method := range methodsToTest {
		ts := getStorer()
		defer ts.disconnect()

		b := bytes.NewBuffer([]byte(`{"pa"` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiCollections(ts), method, testAdminCollectionsURL, b)

		if status != http.StatusBadRequest {
			t.Error("Got:", status, "Expected:", http.StatusBadRequest)
		}
	}
}

func TestHandleAdminTaxiiCollectionsFailServer(t *testing.T) {
	tearDownSQLite()
	ts := getStorer()
	defer ts.disconnect()

	for _, method := range methodsToTest {
		b := bytes.NewBuffer([]byte(`{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiCollections(ts), method, testAdminCollectionsURL, b)

		if status != http.StatusInternalServerError {
			t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
		}
	}
}

/* discovery */

func TestHandleAdminTaxiiDiscoveryInvalidMethod(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleAdminTaxiiDiscovery(ts), "CUSTOM", testAdminDiscoveryURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestHandleAdminTaxiiDiscovery(t *testing.T) {
	setupSQLite()

	tests := []struct {
		method   string
		data     taxiiDiscovery
		expected taxiiDiscovery
	}{
		{method: "DELETE", data: taxiiDiscovery{Title: "deleted"}, expected: taxiiDiscovery{}},
		{method: "POST", data: taxiiDiscovery{Title: "posted"}, expected: taxiiDiscovery{Title: "posted"}},
		{method: "PUT", data: taxiiDiscovery{Title: "updated"}, expected: taxiiDiscovery{Title: "updated"}},
	}

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		payload, err := json.Marshal(test.data)
		if err != nil {
			t.Fatal(err)
		}

		b := bytes.NewBuffer([]byte(payload))
		status, _ := handlerTest(handleAdminTaxiiDiscovery(ts), test.method, testAdminDiscoveryURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK)
		}

		testTd := test.data
		result := taxiiDiscovery{Title: testTd.Title}

		err = result.read(ts)
		if err != nil {
			t.Fatal(err)
		}

		expected := test.expected

		if result.Title != expected.Title {
			t.Error("Title:", result.Title, "Expected:", expected.Title, "Method:", test.method)
		}
	}
}

func TestHandleAdminTaxiiDiscoveryFailAuth(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	b := bytes.NewBuffer([]byte(`{"this won't get processed": true}`))

	for _, method := range methodsToTest {
		for _, test := range contextTests {
			ctx := testingContext()

			for k, v := range test.contexts {
				ctx = context.WithValue(ctx, k, v)
			}

			req := httptest.NewRequest(method, testAdminDiscoveryURL, b)
			req = req.WithContext(ctx)
			res := httptest.NewRecorder()
			handleAdminTaxiiDiscovery(ts)(res, req)

			if res.Code != test.status {
				t.Error("Got:", res.Code, "Expected:", test.status)
			}
		}
	}
}

func TestHandleAdminTaxiiDiscoveryFailData(t *testing.T) {
	setupSQLite()
	methods := []string{"PUT", "POST"}

	for _, method := range methods {
		ts := getStorer()
		defer ts.disconnect()

		b := bytes.NewBuffer([]byte(`{"pa"` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiDiscovery(ts), method, testAdminDiscoveryURL, b)

		if status != http.StatusBadRequest {
			t.Error("Got:", status, "Expected:", http.StatusBadRequest)
		}
	}
}

func TestHandleAdminTaxiiDiscoveryFailServer(t *testing.T) {
	tearDownSQLite()
	ts := getStorer()
	defer ts.disconnect()

	for _, method := range methodsToTest {
		b := bytes.NewBuffer([]byte(`{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiDiscovery(ts), method, testAdminDiscoveryURL, b)

		if status != http.StatusInternalServerError {
			t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
		}
	}
}

/* user */

func TestHandleAdminTaxiiUserInvalidMethod(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleAdminTaxiiUser(ts), "CUSTOM", testAdminUserURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestHandleAdminTaxiiUser(t *testing.T) {
	setupSQLite()

	type taxiiUserTest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		CanAdmin bool   `json:"can_admin"`
	}

	tests := []struct {
		method   string
		data     taxiiUserTest
		expected taxiiUser
	}{
		{method: "DELETE", data: taxiiUserTest{Email: testUser}, expected: taxiiUser{}},
		{method: "POST", data: taxiiUserTest{Email: testUser, Password: testPass, CanAdmin: false}, expected: taxiiUser{Email: testUser, CanAdmin: false}},
		{method: "PUT", data: taxiiUserTest{Email: testUser, CanAdmin: true}, expected: taxiiUser{Email: testUser, CanAdmin: true}},
	}

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		payload, err := json.Marshal(test.data)
		if err != nil {
			t.Fatal(err)
		}

		b := bytes.NewBuffer(payload)
		status, _ := handlerTest(handleAdminTaxiiUser(ts), test.method, testAdminUserURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK, "Method:", test.method)
		}

		testUser := test.data
		result := taxiiUser{Email: testUser.Email}

		err = result.read(ts, hash(testPass))
		if err != nil {
			t.Fatal(err)
		}

		expected := test.expected

		if result.Email != expected.Email {
			t.Error("Email:", result.Email, "Expected:", expected.Email, "Method:", test.method)
		}
		if result.CanAdmin != expected.CanAdmin {
			t.Error("CanAdmin:", result.CanAdmin, "Expected:", expected.CanAdmin, "Method:", test.method)
		}
	}
}

func TestHandleAdminTaxiiUserFailAuth(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	b := bytes.NewBuffer([]byte(`{"this won't get processed": true}`))

	for _, method := range methodsToTest {
		for _, test := range contextTests {
			ctx := testingContext()

			for k, v := range test.contexts {
				ctx = context.WithValue(ctx, k, v)
			}

			req := httptest.NewRequest(method, testAdminUserURL, b)
			req = req.WithContext(ctx)
			res := httptest.NewRecorder()
			handleAdminTaxiiUser(ts)(res, req)

			if res.Code != test.status {
				t.Error("Got:", res.Code, "Expected:", test.status)
			}
		}
	}
}

func TestHandleAdminTaxiiUserFailData(t *testing.T) {
	setupSQLite()

	for _, method := range methodsToTest {
		ts := getStorer()
		defer ts.disconnect()

		b := bytes.NewBuffer([]byte(`{"pa"` + t.Name()))
		status, _ := handlerTest(handleAdminTaxiiUser(ts), method, testAdminUserURL, b)

		if status != http.StatusBadRequest {
			t.Error("Got:", status, "Expected:", http.StatusBadRequest)
		}
	}
}

func TestHandleAdminTaxiiUserFailServer(t *testing.T) {
	tearDownSQLite()
	ts := getStorer()
	defer ts.disconnect()

	for _, method := range methodsToTest {
		b := bytes.NewBuffer([]byte(`{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiUser(ts), method, testAdminUserURL, b)

		if status != http.StatusInternalServerError {
			t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
		}
	}
}

/* user collection */

func TestHandleAdminTaxiiUserCollectionInvalidMethod(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleAdminTaxiiUserCollection(ts), "CUSTOM", testAdminUserCollectionURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestHandleAdminTaxiiUserCollection(t *testing.T) {
	setupSQLite()

	id, err := taxiiIDFromString(testCollectionID)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		method   string
		data     taxiiUserCollection
		expected taxiiUserCollection
	}{
		{method: "DELETE",
			data:     taxiiUserCollection{Email: testUser, taxiiCollectionAccess: taxiiCollectionAccess{ID: id}},
			expected: taxiiUserCollection{Email: "", taxiiCollectionAccess: taxiiCollectionAccess{ID: taxiiID{}, CanRead: false, CanWrite: false}}},
		{method: "POST",
			data:     taxiiUserCollection{Email: testUser, taxiiCollectionAccess: taxiiCollectionAccess{ID: id, CanRead: true, CanWrite: false}},
			expected: taxiiUserCollection{Email: testUser, taxiiCollectionAccess: taxiiCollectionAccess{ID: id, CanRead: true, CanWrite: false}}},
		{method: "PUT",
			data:     taxiiUserCollection{Email: testUser, taxiiCollectionAccess: taxiiCollectionAccess{ID: id, CanRead: true, CanWrite: true}},
			expected: taxiiUserCollection{Email: testUser, taxiiCollectionAccess: taxiiCollectionAccess{ID: id, CanRead: true, CanWrite: true}}},
	}

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		payload, err := json.Marshal(test.data)
		if err != nil {
			t.Fatal(err)
		}

		b := bytes.NewBuffer(payload)
		status, _ := handlerTest(handleAdminTaxiiUserCollection(ts), test.method, testAdminUserCollectionURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK, "Method:", test.method)
		}

		testTuc := test.data
		testTca := testTuc.taxiiCollectionAccess

		resultTuc := taxiiUserCollection{Email: testTuc.Email, taxiiCollectionAccess: testTca}
		err = resultTuc.read(ts)
		if err != nil {
			t.Fatal(err)
		}
		resultTca := resultTuc.taxiiCollectionAccess

		expectedTuc := test.expected
		expectedTca := expectedTuc.taxiiCollectionAccess

		if resultTuc.Email != expectedTuc.Email {
			t.Error("Got:", resultTuc.Email, "Expected:", expectedTuc.Email, "Method:", test.method)
		}
		if resultTca.ID != expectedTca.ID {
			t.Error("Got:", resultTca.ID.String(), "Expected:", expectedTca.ID.String(), "Method:", test.method)
		}
		if resultTca.CanRead != expectedTca.CanRead {
			t.Error("Got:", resultTca.CanRead, "Expected:", expectedTca.CanRead, "Method:", test.method)
		}
		if resultTca.CanWrite != expectedTca.CanWrite {
			t.Error("Got:", resultTca.CanWrite, "Expected:", expectedTca.CanWrite, "Method:", test.method)
		}
	}
}

func TestHandleAdminTaxiiUserCollectionFailAuth(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	b := bytes.NewBuffer([]byte(`{"this won't get processed": true}`))

	for _, method := range methodsToTest {
		for _, test := range contextTests {
			ctx := testingContext()

			for k, v := range test.contexts {
				ctx = context.WithValue(ctx, k, v)
			}

			req := httptest.NewRequest(method, testAdminUserCollectionURL, b)
			req = req.WithContext(ctx)
			res := httptest.NewRecorder()
			handleAdminTaxiiUserCollection(ts)(res, req)

			if res.Code != test.status {
				t.Error("Got:", res.Code, "Expected:", test.status)
			}
		}
	}
}

func TestHandleAdminTaxiiUserCollectionFailData(t *testing.T) {
	setupSQLite()

	for _, method := range methodsToTest {
		ts := getStorer()
		defer ts.disconnect()

		b := bytes.NewBuffer([]byte(`{"pa"` + t.Name()))
		status, _ := handlerTest(handleAdminTaxiiUserCollection(ts), method, testAdminUserCollectionURL, b)

		if status != http.StatusBadRequest {
			t.Error("Got:", status, "Expected:", http.StatusBadRequest)
		}
	}
}

func TestHandleAdminTaxiiUserCollectionFailServer(t *testing.T) {
	tearDownSQLite()

	ts := getStorer()
	defer ts.disconnect()

	for _, method := range methodsToTest {
		b := bytes.NewBuffer([]byte(`{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiUserCollection(ts), method, testAdminUserCollectionURL, b)

		if status != http.StatusInternalServerError {
			t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
		}
	}
}

/* user password */

func TestHandleAdminTaxiiUserPasswordInvalidMethod(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleAdminTaxiiUserPassword(ts), "CUSTOM", testAdminUserPasswordURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestHandleAdminTaxiiUserPassword(t *testing.T) {
	setupSQLite()

	tests := []struct {
		method   string
		data     taxiiUserPassword
		expected taxiiUser
	}{
		{method: "PUT",
			data:     taxiiUserPassword{Email: testUser, Password: "updated"},
			expected: taxiiUser{Email: testUser, CanAdmin: true}},
	}

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		payload, err := json.Marshal(test.data)
		if err != nil {
			t.Fatal(err)
		}

		b := bytes.NewBuffer([]byte(payload))
		status, _ := handlerTest(handleAdminTaxiiUserPassword(ts), test.method, testAdminUserPasswordURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK, "Method:", test.method)
		}

		testTup := test.data
		result := taxiiUser{Email: testTup.Email}
		err = result.read(ts, hash(testTup.Password))
		if err != nil {
			t.Fatal(err)
		}

		expected := test.expected

		if result.Email != expected.Email {
			t.Error("Got:", result.Email, "Expected:", expected.Email, "Method:", test.method)
		}
		if result.CanAdmin != expected.CanAdmin {
			t.Error("Got:", result.CanAdmin, "Expected:", expected.CanAdmin, "Method:", test.method)
		}
	}
}

func TestHandleAdminTaxiiUserPasswordFailAuth(t *testing.T) {
	setupSQLite()
	methods := []string{"PUT"}

	ts := getStorer()
	defer ts.disconnect()

	b := bytes.NewBuffer([]byte(`{"this won't get processed": true}`))

	for _, method := range methods {
		for _, test := range contextTests {
			ctx := testingContext()

			for k, v := range test.contexts {
				ctx = context.WithValue(ctx, k, v)
			}

			req := httptest.NewRequest(method, testAdminUserPasswordURL, b)
			req = req.WithContext(ctx)
			res := httptest.NewRecorder()
			handleAdminTaxiiUserPassword(ts)(res, req)

			if res.Code != test.status {
				t.Error("Got:", res.Code, "Expected:", test.status)
			}
		}
	}
}

func TestHandleAdminTaxiiUserPasswordFailData(t *testing.T) {
	setupSQLite()
	methods := []string{"PUT"}

	ts := getStorer()
	defer ts.disconnect()

	for _, method := range methods {
		b := bytes.NewBuffer([]byte(`{"pa"` + t.Name()))
		status, _ := handlerTest(handleAdminTaxiiUserPassword(ts), method, testAdminUserPasswordURL, b)

		if status != http.StatusBadRequest {
			t.Error("Got:", status, "Expected:", http.StatusBadRequest)
		}
	}
}

func TestHandleAdminTaxiiUserPasswordFailServer(t *testing.T) {
	tearDownSQLite()
	methods := []string{"PUT"}

	ts := getStorer()
	defer ts.disconnect()

	for _, method := range methods {
		b := bytes.NewBuffer([]byte(`{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiUserPassword(ts), method, testAdminUserPasswordURL, b)

		if status != http.StatusInternalServerError {
			t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
		}
	}
}
