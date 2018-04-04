package main

import s "github.com/pladdy/stones"

type taxiiManifest struct {
	Objects []taxiiManifestEntry `json:"objects"`
}

type taxiiManifestEntry struct {
	ID         s.StixID `json:"id"`
	DateAdded  string   `json:"date_added"`
	Versions   []string `json:"versions"`
	MediaTypes []string `json:"media_types"`
}
