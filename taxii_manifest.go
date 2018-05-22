package main

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func handleTaxiiManifest(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := taxiiManifest{}

		tr, err := newTaxiiRange(r.Header.Get("Range"))
		if err != nil {
			rangeNotSatisfiable(w, err)
			return
		}
		r = withTaxiiRange(r, tr)

		tf := newTaxiiFilter(r)
		result, err := m.read(ts, takeCollectionID(r), tf)
		if err != nil {
			log.WithFields(
				log.Fields{"fn": "handleTaxiiManifest", "id": takeCollectionID(r), "error": err},
			).Error("failed to read manifest")
			resourceNotFound(w, errors.New("Unable to get manifest"))
			return
		}

		result.data = withStixContentType(result.data.(taxiiManifest))

		if tr.Valid() {
			tr.total = result.items
			w.Header().Set("Content-Range", tr.String())
			writePartialContent(w, taxiiContentType, resourceToJSON(result.data))
		} else {
			writeContent(w, taxiiContentType, resourceToJSON(result.data))
		}
	})
}

func withStixContentType(m taxiiManifest) taxiiManifest {
	for i := range m.Objects {
		m.Objects[i].MediaTypes = []string{stixContentType}
	}
	return m
}

type taxiiManifest struct {
	Objects []taxiiManifestEntry `json:"objects"`
}

func (t *taxiiManifest) read(ts taxiiStorer, collectionID string, tf taxiiFilter) (taxiiResult, error) {
	tm := *t

	result, err := ts.read("taxiiManifest", []interface{}{collectionID}, tf)
	if err != nil {
		return result, err
	}

	tm = result.data.(taxiiManifest)
	*t = tm
	return result, err
}

type taxiiManifestEntry struct {
	ID         string   `json:"id"`
	DateAdded  string   `json:"date_added"`
	Versions   []string `json:"versions"`
	MediaTypes []string `json:"media_types"`
}
