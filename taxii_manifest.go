package main

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type taxiiManifest struct {
	Objects []taxiiManifestEntry `json:"objects"`
}

func (t *taxiiManifest) read(ts taxiiStorer, collectionID string) error {
	tm := *t

	result, err := ts.read("taxiiManifest", []interface{}{collectionID})
	if err != nil {
		return err
	}

	tm = result.(taxiiManifest)
	*t = tm
	return err
}

func handleTaxiiManifest(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := taxiiManifest{}

		err := m.read(ts, getCollectionID(r.URL.Path))
		if err != nil {
			log.WithFields(
				log.Fields{"fn": "handleTaxiiManifest", "id": getCollectionID(r.URL.Path), "error": err},
			).Error("failed to read manifest")

			resourceNotFound(w, errors.New("Unable to get manifest"))
			return
		}

		for i := range m.Objects {
			m.Objects[i].MediaTypes = []string{stixContentType}
		}

		writeContent(w, taxiiContentType, resourceToJSON(m))
	})
}

type taxiiManifestEntry struct {
	ID         string   `json:"id"`
	DateAdded  string   `json:"date_added"`
	Versions   []string `json:"versions"`
	MediaTypes []string `json:"media_types"`
}
