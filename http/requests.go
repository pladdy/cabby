package http

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

const (
	defaultVersion     = "last"
	jsonContentType    = "application/json"
	sixMonthsOfSeconds = "63072000"
)

// I don't want to introduce a router dependency, so I'm using regexes...which may be worse.  I felt this was a less
// bad path.
// Golang regexes: https://github.com/google/re2/wiki/Syntax
// Regex testing tool: https://regex101.com/ (the fact I needed this may indicate this was a bad choice)
var (
	apiRootRegex         = `/(?P<apiroot>[\w\\\/]+)/`
	apiRootPathRegex     = regexp.MustCompile(apiRootRegex + `(status|collections)/`)
	statusPathRegex      = regexp.MustCompile(apiRootRegex + `status/(?P<statusid>[a-zA-Z\-\d]+)/?`)
	collectionsPathRegex = regexp.MustCompile(apiRootRegex + "collections/")
	collectionPathRegex  = regexp.MustCompile(collectionsPathRegex.String() + `(?P<collectionid>[a-zA-Z\-\d]+)/?`)
	manifestPathRegex    = regexp.MustCompile(collectionPathRegex.String() + "/manifest/")
	objectsPathRegex     = regexp.MustCompile(collectionPathRegex.String() + "/objects/")
	objectPathRegex      = regexp.MustCompile(objectsPathRegex.String() + `(?P<objectid>[a-zA-Z\-\d]+)/?`)
	versionsPathRegex    = regexp.MustCompile(objectPathRegex.String() + "versions/?")
)

func newFilter(r *http.Request) (f cabby.Filter) {
	f.AddedAfter = takeAddedAfter(r)
	f.IDs = takeMatchIDs(r)
	f.Types = takeMatchTypes(r)

	f.Versions = takeMatchVersions(r)
	if f.Versions == "" {
		f.Versions = defaultVersion
	}
	return
}

func takeAddedAfter(r *http.Request) stones.Timestamp {
	af := r.URL.Query()["added_after"]

	if len(af) > 0 {
		t, err := stones.TimestampFromString(af[0])
		if err == nil {
			return t
		}
	}
	return stones.Timestamp{}
}

func takeAPIRoot(r *http.Request) string {
	apiRootIndex := 1
	if apiRootPathRegex.Match([]byte(r.URL.Path)) {
		return apiRootPathRegex.FindStringSubmatch(r.URL.Path)[apiRootIndex]
	}
	return ""
}

func takeCollectionAccess(r *http.Request) cabby.CollectionAccess {
	u := cabby.TakeUser(r.Context())
	ca := u.CollectionAccessList

	id, err := cabby.IDFromString(takeCollectionID(r))
	if err != nil {
		log.WithFields(log.Fields{
			"user":         u.Email,
			"collectionID": takeCollectionID(r),
			"error":        err,
		}).Warn("Failed to convert collection id from string")
		return cabby.CollectionAccess{}
	}

	return ca[id]
}

func takeCollectionID(r *http.Request) string {
	collectionIndex := 2
	if collectionPathRegex.Match([]byte(r.URL.Path)) {
		return collectionPathRegex.FindStringSubmatch(r.URL.Path)[collectionIndex]
	}
	return ""
}

func takeLimit(r *http.Request) string {
	ls := r.URL.Query()["limit"]

	if len(ls) > 0 {
		return ls[0]
	}
	return ""
}

func takeMatchFilters(r *http.Request, filter string) string {
	filters := r.URL.Query()[filter]

	if len(filters) > 0 {
		return strings.Join(filters, ",")
	}
	return ""
}

func takeMatchIDs(r *http.Request) string {
	return takeMatchFilters(r, "match[id]")
}

func takeMatchTypes(r *http.Request) string {
	return takeMatchFilters(r, "match[type]")
}

func takeMatchVersions(r *http.Request) string {
	return takeMatchFilters(r, "match[version]")
}

func takeObjectID(r *http.Request) string {
	objectIndex := 3
	if objectPathRegex.Match([]byte(r.URL.Path)) {
		return objectPathRegex.FindStringSubmatch(r.URL.Path)[objectIndex]
	}
	return ""
}

func trimSlashes(s string) string {
	return strings.Trim(s, "/")
}

func takeStatusID(r *http.Request) string {
	statusIndex := 2
	if statusPathRegex.Match([]byte(r.URL.Path)) {
		return statusPathRegex.FindStringSubmatch(r.URL.Path)[statusIndex]
	}
	return ""
}

func takeVersions(r *http.Request) string {
	if versionsPathRegex.Match([]byte(r.URL.Path)) {
		return "versions"
	}
	return ""
}

func withTransactionID(r *http.Request) *http.Request {
	transactionID := uuid.Must(uuid.NewV4())
	return r.WithContext(cabby.WithTransactionID(r.Context(), transactionID))
}
