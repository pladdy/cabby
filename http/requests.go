package http

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

const (
	defaultVersion     = "last"
	jsonContentType    = "application/json"
	sixMonthsOfSeconds = "63072000"
)

func getToken(s string, i int) string {
	tokens := strings.Split(s, "/")

	if len(tokens) > i {
		return tokens[i]
	}
	return ""
}

func lastURLPathToken(u string) string {
	u = strings.TrimSuffix(u, "/")
	tokens := strings.Split(u, "/")
	return tokens[len(tokens)-1]
}

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

func takeAddedAfter(r *http.Request) string {
	af := r.URL.Query()["added_after"]

	if len(af) > 0 {
		t, err := time.Parse(time.RFC3339Nano, af[0])
		if err == nil {
			return t.Format(time.RFC3339Nano)
		}
	}
	return ""
}

func takeAPIRoot(r *http.Request) string {
	var apiRootIndex = 1
	return getToken(r.URL.Path, apiRootIndex)
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
	var collectionIndex = 3
	return getToken(r.URL.Path, collectionIndex)
}

func takeObjectID(r *http.Request) string {
	var objectIDIndex = 5
	return getToken(r.URL.Path, objectIDIndex)
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

func trimSlashes(s string) string {
	re := regexp.MustCompile("^/")
	s = re.ReplaceAllString(s, "")

	re = regexp.MustCompile("/$")
	s = re.ReplaceAllString(s, "")

	parts := strings.Split(s, "/")
	return strings.Join(parts, "/")
}

func takeStatusID(r *http.Request) string {
	var statusIndex = 3
	return getToken(r.URL.Path, statusIndex)
}

func withTransactionID(r *http.Request) *http.Request {
	transactionID := uuid.Must(uuid.NewV4())
	return r.WithContext(cabby.WithTransactionID(r.Context(), transactionID))
}

// func userAuthorized(w http.ResponseWriter, r *http.Request) bool {
// 	if !userExists(r) {
// 		unauthorized(w, errors.New("No user specified"))
// 		return false
// 	}
//
// 	if !takeCanAdmin(r) {
// 		forbidden(w, errors.New("Not authorized to create API Roots"))
// 		return false
// 	}
// 	return true
// }
