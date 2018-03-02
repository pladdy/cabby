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

	maxContent := int64(2048)
	b := bytes.NewBuffer(bundle)
	status, response := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", u, b)

	if status != 200 {
		t.Error("Got:", status, "Expected: 200")
	}

	info.Println("Response:", response)

	s := getSQLiteDB()
	defer s.disconnect()

	expectedCount := 0
	var count int
	err := s.db.QueryRow("select count(*) from stix_object where collection_id = '" + testID + "'").Scan(&count)
	if err != nil {
		t.Error(err)
	}

	if count != expectedCount {
		t.Error("Got:", count, "Expected:", expectedCount)
	}
}

func TestHandleTaxiiObjectsPostContentTooBig(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	bundleFile, _ := os.Open("test/data/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(1)
	b := bytes.NewBuffer(bundle)
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", u, b)

	if status != 413 {
		t.Error("Got:", status, "Expected: 413")
	}
}

func TestHandleTaxiiObjectPostCollectionUnauthorized(t *testing.T) {

}
