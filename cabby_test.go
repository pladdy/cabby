package cabby

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

type errorLog struct {
	Error string
	Level string
	Msg   string
}

func TestAPIRootValidate(t *testing.T) {
	tests := []struct {
		apiRoot     APIRoot
		expectError bool
	}{
		{APIRoot{Path: "foo"}, true},
		{APIRoot{}, true},
		{APIRoot{Path: "foo", Title: "title"}, true},
		{APIRoot{Path: "foo", Title: "title", Versions: []string{"taxii-2.0"}}, true},
		{APIRoot{Path: "foo", Title: "title", Versions: []string{"taxii-2.1"}}, false},
		{APIRoot{Path: "foo", Title: "title", Versions: []string{TaxiiVersion}}, false},
	}

	for _, test := range tests {
		result := test.apiRoot.Validate()

		if test.expectError && result == nil {
			t.Error("Testing API Root:", test.apiRoot, "Got an error?", result, "Expected an error?", test.expectError)
		}
	}
}

func TestAPIRootIncludesMinVersion(t *testing.T) {
	tests := []struct {
		apiRoot  APIRoot
		expected bool
	}{
		{APIRoot{Path: "foo"}, false},
		{APIRoot{}, false},
		{APIRoot{Path: "foo", Title: "title"}, false},
		{APIRoot{Versions: []string{TaxiiVersion}}, true},
		{APIRoot{Versions: []string{TaxiiVersion, TaxiiVersion}}, true},
		{APIRoot{Versions: []string{TaxiiVersion, "taxii-2.1"}}, true},
	}

	for _, test := range tests {
		result := test.apiRoot.IncludesMinVersion(test.apiRoot.Versions)

		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestConfigParse(t *testing.T) {
	c := Config{}.Parse("config/cabby.example.json")

	if c.Host != "localhost" {
		t.Error("Got:", "localhost", "Expected:", "localhost")
	}
	if c.Port != 1234 {
		t.Error("Got:", strconv.Itoa(1234), "Expected:", strconv.Itoa(1234))
	}
}

func TestParseConfigNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered", r)
		}
	}()

	_ = Config{}.Parse("foo/bar")
	t.Error("Failed to panic with an unknown resource")
}

func TestConfigParseInvalidJSON(t *testing.T) {
	invalidJSON := "invalid.json"

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered", r)
			os.Remove(invalidJSON)
		}
	}()

	ioutil.WriteFile(invalidJSON, []byte("invalid"), 0644)
	Config{}.Parse(invalidJSON)
	t.Error("Failed to panic with an unknown resource")
}

func TestDiscoveryValidate(t *testing.T) {
	tests := []struct {
		discovery   Discovery
		expectError bool
	}{
		{Discovery{Title: "foo"}, false},
		{Discovery{}, true},
	}

	for _, test := range tests {
		result := test.discovery.Validate()

		if test.expectError && result == nil {
			t.Error("Got:", result, "Expected:", test.expectError)
		}
	}
}

func TestNewCollection(t *testing.T) {
	tests := []struct {
		idString    string
		shouldError bool
	}{
		{"invalid", true},
		{uuid.Must(uuid.NewV4()).String(), false},
	}

	for _, test := range tests {
		c, err := NewCollection(test.idString)

		if test.shouldError && err == nil {
			t.Error("Test with id of", test.idString, "should produce an error!")
		}

		if err == nil && c.ID.String() != test.idString {
			t.Error("Got:", c.ID.String(), "Expected:", test.idString)
		}
	}

	// test if 'collections' is passed; return a uuid
	_, err := NewCollection("collections")
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}
}

func TestCollectionValidate(t *testing.T) {
	validID, _ := NewID()
	validTitle := "a title"

	tests := []struct {
		collection  Collection
		expectError bool
	}{
		{Collection{ID: validID, Title: validTitle}, false},
		{Collection{Title: validTitle}, true},
		{Collection{ID: validID}, true},
		{Collection{}, true},
	}

	for _, test := range tests {
		result := test.collection.Validate()

		if test.expectError && result == nil {
			t.Error("Got:", result, "Expected:", test.expectError)
		}

		if !test.expectError && result != nil {
			t.Error("Error unexpected:", result, "Test:", test)
		}
	}
}

func TestNewPage(t *testing.T) {
	tests := []struct {
		input       string
		resultPage  Page
		expectError bool
	}{
		{"10", Page{Limit: 10}, false},
		{" 10", Page{Limit: 10}, true},
		{"10 ", Page{Limit: 10}, true},
		{"", Page{}, true},
	}

	for _, test := range tests {
		result, err := NewPage(test.input)
		if result != test.resultPage {
			t.Error("Got:", result, "Expected:", test.resultPage)
		}

		if err != nil && test.expectError == false {
			t.Error("Got:", err, "Expected: no error", "Page:", result)
		}
	}
}

func TestPageAddedAfterFirst(t *testing.T) {
	now := stones.NewTimestamp()

	tests := []struct {
		p        Page
		expected string
	}{
		{Page{}, time.Time{}.Format(time.RFC3339Nano)},
		{Page{MinimumAddedAfter: now}, now.String()},
	}

	for _, test := range tests {
		result := test.p.AddedAfterFirst()
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestPageAddedAfterLast(t *testing.T) {
	now := stones.NewTimestamp()

	tests := []struct {
		p        Page
		expected string
	}{
		{Page{}, time.Time{}.Format(time.RFC3339Nano)},
		{Page{MaximumAddedAfter: now}, now.String()},
	}

	for _, test := range tests {
		result := test.p.AddedAfterLast()
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestPageSetAddedAfters(t *testing.T) {
	p := Page{}

	// a page is mutated when setting a date
	// it will be used as a returned data set is iterated over

	// the first date passed to a page should set min and max times
	first, _ := time.Parse(time.RFC3339Nano, "2016-02-01T00:00:01.123Z")
	p.SetAddedAfters(first.Format(time.RFC3339Nano))

	if p.MinimumAddedAfter.UnixNano() != first.UnixNano() {
		t.Error("Got:", p.MinimumAddedAfter.UnixNano(), "Expected:", first.UnixNano())
	}
	if p.MaximumAddedAfter.UnixNano() != first.UnixNano() {
		t.Error("Got:", p.MinimumAddedAfter.UnixNano(), "Expected:", first.UnixNano())
	}

	// the second date is a duplicate, it won't change anything
	second, _ := time.Parse(time.RFC3339Nano, "2016-02-01T00:00:01.123Z")
	p.SetAddedAfters(second.Format(time.RFC3339Nano))

	if p.MinimumAddedAfter.UnixNano() != second.UnixNano() {
		t.Error("Got:", p.MinimumAddedAfter.UnixNano(), "Expected:", second.UnixNano())
	}
	if p.MaximumAddedAfter.UnixNano() != second.UnixNano() {
		t.Error("Got:", p.MinimumAddedAfter.UnixNano(), "Expected:", second.UnixNano())
	}

	third, _ := time.Parse(time.RFC3339Nano, "2017-02-01T00:00:01.123Z")
	p.SetAddedAfters(third.Format(time.RFC3339Nano))

	if p.MinimumAddedAfter.UnixNano() != first.UnixNano() {
		t.Error("Got:", p.MinimumAddedAfter.UnixNano(), "Expected:", first.UnixNano())
	}
	if p.MaximumAddedAfter.UnixNano() != third.UnixNano() {
		t.Error("Got:", p.MinimumAddedAfter.UnixNano(), "Expected:", third.UnixNano())
	}

	fourth, _ := time.Parse(time.RFC3339Nano, "2015-02-01T00:00:01.123Z")
	p.SetAddedAfters(fourth.Format(time.RFC3339Nano))

	if p.MinimumAddedAfter.UnixNano() != fourth.UnixNano() {
		t.Error("Got:", p.MinimumAddedAfter.UnixNano(), "Expected:", fourth.UnixNano())
	}
	if p.MaximumAddedAfter.UnixNano() != third.UnixNano() {
		t.Error("Got:", p.MinimumAddedAfter.UnixNano(), "Expected:", third.UnixNano())
	}
}

func TestPageSetAddedAftersFail(t *testing.T) {
	// redirect log output for test
	var buf bytes.Buffer

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(&buf)

	defer func() {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stderr)
	}()

	p := Page{}
	p.SetAddedAfters("fail")

	// parse log into struct
	var result errorLog
	err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &result)
	if err != nil {
		t.Fatal(err)
	}

	level := "error"
	if result.Level != level {
		t.Error("Got:", result.Level, "Expected:", level)
	}
}

func TestPageValid(t *testing.T) {
	tests := []struct {
		testPage Page
		expected bool
	}{
		{Page{Limit: 10}, true},
		{Page{Limit: 0}, false},
	}

	for _, test := range tests {
		result := test.testPage.Valid()
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestNewID(t *testing.T) {
	_, err := NewID()
	if err != nil {
		t.Error("Expected no error:", err)
	}
}

func TestIDFromString(t *testing.T) {
	uid := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	id, err := IDFromString(uid)
	if err != nil {
		t.Error("Expected no error:", err)
	}

	if id.String() != uid {
		t.Error("Got:", id.String(), "Expected:", uid)
	}
}

func TestIDFromStringBadInput(t *testing.T) {
	_, err := IDFromString("")
	if err == nil {
		t.Error("Expected an error")
	}

	_, err = IDFromString("fail")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestIDUsingString(t *testing.T) {
	id, err := IDUsingString("")
	if err != nil {
		t.Fatal(err)
	}

	expected := "101afa45-b9bd-5b31-8734-0a59e5cc3db3"
	if id.String() != expected {
		t.Error("Got:", id.String(), "Expected:", expected)
	}
}

func TestIDIsEmpty(t *testing.T) {
	id, err := NewID()
	if err != nil {
		t.Fatal(err)
	}

	if id.IsEmpty() == true {
		t.Error("Expected to NOT be empty")
	}

	emptyID := ID{}
	if emptyID.IsEmpty() == false {
		t.Error("Expected ID to be empty")
	}
}

func TestNewStatus(t *testing.T) {
	_, err := NewStatus(1)
	if err != nil {
		t.Error("Got:", err, "Expected: no error")
	}
}

func TestNewStatusError(t *testing.T) {
	_, err := NewStatus(0)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestUserDefined(t *testing.T) {
	tests := []struct {
		user     User
		expected bool
	}{
		{user: User{Email: "foo"}, expected: true},
		{user: User{}, expected: false},
	}

	for _, test := range tests {
		result := test.user.Defined()
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestUserValidate(t *testing.T) {
	tests := []struct {
		user        User
		expectError bool
	}{
		{User{Email: "foo"}, true},
		{User{}, true},
		{User{Email: "no@no.no"}, false},
		{User{Email: "some-person@yaoo.co.uk"}, false},
	}

	for _, test := range tests {
		result := test.user.Validate()

		if test.expectError && result == nil {
			t.Error("Got:", result, "Expected:", test.expectError)
		}
	}
}
