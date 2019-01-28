package sqlite

import (
	"regexp"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
)

func TestNewDataStore(t *testing.T) {
	_, err := NewDataStore("temp.db")

	if err != nil {
		t.Error("Got:", err, "Expected: nil")
	}
}

func TestNewDataStoreNoPath(t *testing.T) {
	_, err := NewDataStore("")

	if err == nil {
		t.Error("Expected an error")
	}
}

func TestDataStoreClose(t *testing.T) {
	s, err := NewDataStore("temp.db")

	if err != nil {
		t.Error("Got:", err, "Expected: nil")
	}

	s.Close()
}

func TestSQLiteBatchWriteSmall(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	sql := `insert into collection (id, api_root_path, title, description, media_types)
					values (?, ?, ?, ?, ?)`

	go ds.batchWrite(sql, toWrite, errs)
	toWrite <- []interface{}{"test", "test api root", "test collection", "this is a test collection", "media type"}
	close(toWrite)

	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}

	var id string
	err := ds.DB.QueryRow(`select id from collection where id = 'test'`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	if id != "test" {
		t.Error("Got:", id, "Expected:", "test")
	}
}

func TestSQLiteBatchWriteExecuteLarge(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	sql := `insert into collection (id, api_root_path, title, description)
					values (?, ?, ?, ?)`

	go ds.batchWrite(sql, toWrite, errs)

	recordsToWrite := 1000
	for i := 0; i <= recordsToWrite; i++ {
		toWrite <- []interface{}{"test" + string(i), "api root", "collection", "a test collection"}
	}
	close(toWrite)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError != nil {
		t.Error("Got:", lastError, "Expected: no error")
	}
}

func TestSQLiteBatchWriteWriteOperationError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	go ds.batchWrite("fail", toWrite, errs)
	toWrite <- []interface{}{"fail"}
	close(toWrite)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError == nil {
		t.Error("Expected error")
	}
}

func TestSQLiteBatchWriteExecuteError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	sql := `insert into collection (id, api_root_path, title, description)
					values (?, ?, ?, ?)`

	go ds.batchWrite(sql, toWrite, errs)

	for i := 0; i <= maxWritesPerBatch; i++ {
		if i == maxWritesPerBatch {
			// a commit is about to happen
			ds.Close()
		}
		toWrite <- []interface{}{"test" + string(i), "api root", "collection", "a test collection"}
	}
	close(toWrite)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError == nil {
		t.Error("Expected error")
	}
}

func TestSQLiteBatchWriteCommitError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	go ds.batchWrite("insert into objects (id, object) values (?, ?)", toWrite, errs)
	toWrite <- []interface{}{"fail"}
	close(toWrite)

	var lastError error
	for e := range errs {
		if e != nil {
			lastError = e
		}
	}

	if lastError == nil {
		t.Error("Expected error")
	}
}

func TestDataStoreWriteError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	err := ds.write("this is not a valid query")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestDataStoreExecuteError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	tx, _ := ds.DB.Begin()
	stmt, _ := tx.Prepare("select * from objects")

	err := ds.execute(stmt, "fail")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestDataStoreWriteOperationError(t *testing.T) {
	ds := testDataStore()

	ds.Open()
	_, _, err := ds.writeOperation("this is not a valid query")
	if err == nil {
		t.Error("Expected an error")
	}

	ds.Close()
	_, _, err = ds.writeOperation("this is not a valid query")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestFilterQueryString(t *testing.T) {
	ts, err := stones.TimestampFromString("2016-04-06T20:03:48.123Z")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		filter        cabby.Filter
		expectedQuery string
		expectedArgs  []interface{}
	}{
		{cabby.Filter{}, ``, []interface{}{}},
		{cabby.Filter{AddedAfter: ts},
			`(strftime('%s', strftime('%Y-%m-%d %H:%M:%f', created_at)) + strftime('%f',
				strftime('%Y-%m-%d %H:%M:%f', created_at))) * 1000 > (strftime('%s',
				strftime('%Y-%m-%d %H:%M:%f', ?)) + strftime('%f', strftime('%Y-%m-%d %H:%M:%f', ?))) * 1000`,
			[]interface{}{"2016-04-06T20:03:48.123Z", "2016-04-06T20:03:48.123Z"}},
		{cabby.Filter{IDs: "indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f"},
			"(id = ?)",
			[]interface{}{"indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f"}},
		{cabby.Filter{IDs: "indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f,malware--31b940d4-6f7f-459a-80ea-9c1f17b5891b"},
			"(id = ? or id = ?)",
			[]interface{}{"indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f", "malware--31b940d4-6f7f-459a-80ea-9c1f17b5891b"}},
		{cabby.Filter{Types: "indicator"},
			"(type = ?)",
			[]interface{}{"indicator"}},
		{cabby.Filter{Types: "indicator,malware"},
			"(type = ? or type = ?)",
			[]interface{}{"indicator", "malware"}},
		{cabby.Filter{Versions: "2018-10-30T12:03:48.123Z"},
			`((strftime('%s', strftime('%Y-%m-%d %H:%M:%f', modified)) + strftime('%f',
				 strftime('%Y-%m-%d %H:%M:%f', modified))) * 1000 = (strftime('%s',
				 strftime('%Y-%m-%d %H:%M:%f', ?)) + strftime('%f', strftime('%Y-%m-%d %H:%M:%f', ?))) * 1000)`,
			[]interface{}{"2018-10-30T12:03:48.123Z", "2018-10-30T12:03:48.123Z"}},
		{cabby.Filter{Versions: "2016-04-06T20:03:48.123Z,2016-04-07T20:03:48.123Z"},
			`((strftime('%s', strftime('%Y-%m-%d %H:%M:%f', modified)) + strftime('%f',
				 strftime('%Y-%m-%d %H:%M:%f', modified))) * 1000 = (strftime('%s',
				 strftime('%Y-%m-%d %H:%M:%f', ?)) + strftime('%f', strftime('%Y-%m-%d %H:%M:%f', ?))) * 1000
        or (strftime('%s', strftime('%Y-%m-%d %H:%M:%f', modified)) + strftime('%f',
				 strftime('%Y-%m-%d %H:%M:%f', modified))) * 1000 = (strftime('%s',
				 strftime('%Y-%m-%d %H:%M:%f', ?)) + strftime('%f', strftime('%Y-%m-%d %H:%M:%f', ?))) * 1000)`,
			[]interface{}{"2016-04-06T20:03:48.123Z", "2016-04-06T20:03:48.123Z",
				"2016-04-07T20:03:48.123Z", "2016-04-07T20:03:48.123Z"}},
		{cabby.Filter{Versions: "first"},
			"(version in ('first', 'only'))",
			[]interface{}{}},
		{cabby.Filter{Versions: "last"},
			"(version in ('last', 'only'))",
			[]interface{}{}},
		{cabby.Filter{Versions: "all"},
			"(1 = 1)",
			[]interface{}{}},
	}

	for _, test := range tests {
		filter := Filter{test.filter}
		resultQuery, resultArgs := filter.QueryString()

		// condense the whitespace to
		re := regexp.MustCompile(`\s+`)
		resultQuery = re.ReplaceAllString(resultQuery, " ")
		test.expectedQuery = re.ReplaceAllString(test.expectedQuery, " ")

		if resultQuery != test.expectedQuery {
			t.Error("Got:", resultQuery, "Expected:", test.expectedQuery)
		}

		if len(resultArgs) != len(test.expectedArgs) {
			t.Error("Got:", len(resultArgs), "Expected:", len(test.expectedArgs))
		}

		for i := 0; i < len(resultArgs); i++ {
			if resultArgs[i] != test.expectedArgs[i] {
				t.Error("Got:", resultArgs[i], "Expected:", test.expectedArgs[i])
			}
		}
	}
}

func TestRangeQueryString(t *testing.T) {
	tests := []struct {
		r     Range
		query string
		args  int
	}{
		{Range{&cabby.Range{}}, "", 0},
		{Range{&cabby.Range{First: 0, Last: 0, Set: true}}, "limit ? offset ?", 2},
	}

	for _, test := range tests {
		result, args := test.r.QueryString()

		if result != test.query {
			t.Error("Got:", result, "Expected:", test.query)
		}

		if len(args) != test.args {
			t.Error("Got:", len(args), "Expected:", test.args)
		}
	}
}
