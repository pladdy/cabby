package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestHandleTaxiiObjectsPost(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	bundleFile, _ := os.Open("test/data/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)

	b := bytes.NewBuffer(bundle)
	status, response := handlerTest(handleTaxiiObjects(ts), "POST", u, b)

	if status != 200 {
		t.Error("Got:", status, "Expected: 200")
	}

	info.Println("Response:", response)

	s := getSQLiteDB()
	defer s.disconnect()

	expectedCount := 3
	var count int
	err := s.db.QueryRow("select count(*) from stix_objects where collection_id = '" + testID + "'").Scan(&count)
	if err != nil {
		t.Error(err)
	}

	if count != expectedCount {
		t.Error("Got:", count, "Expected:", expectedCount)
	}
}

func TestHandleTaxiiObjectsPostContentTooBig(t *testing.T) {

}

func TestHandleTaxiiObjectPostCollectionUnauthorized(t *testing.T) {

}
