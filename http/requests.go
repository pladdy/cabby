package http

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"
)

// per context docuentation, use a key type for context keys
type key int

const (
	keyCanAdmin             key = 0
	keyCollectionAccessList key = 1
	keyRequestRange         key = 2
	keyTransactionID        key = 3
	keyUserName             key = 4
	defaultVersion              = "last"
	jsonContentType             = "application/json"
	sixMonthsOfSeconds          = "63072000"
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

// func takeCanAdmin(r *http.Request) bool {
// 	ca, ok := r.Context().Value(canAdmin).(bool)
// 	if !ok {
// 		return false
// 	}
// 	return ca
// }

func takeCollectionAccess(r *http.Request) cabby.CollectionAccess {
	// get collection access map from context
	ca, ok := r.Context().Value(keyCollectionAccessList).(map[cabby.ID]cabby.CollectionAccess)
	if !ok {
		log.WithFields(log.Fields{
			"user":         takeUser(r),
			"collectionID": takeCollectionID(r),
		}).Warn("Failed to get collection access from context")
		return cabby.CollectionAccess{}
	}

	id, err := cabby.IDFromString(takeCollectionID(r))
	if err != nil {
		log.WithFields(log.Fields{
			"user":         takeUser(r),
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

func takeStatusID(r *http.Request) string {
	var statusIndex = 3
	return getToken(r.URL.Path, statusIndex)
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

func takeUser(r *http.Request) string {
	user, ok := r.Context().Value(keyUserName).(string)
	if !ok {
		return ""
	}
	return user
}

func trimSlashes(s string) string {
	re := regexp.MustCompile("^/")
	s = re.ReplaceAllString(s, "")

	re = regexp.MustCompile("/$")
	s = re.ReplaceAllString(s, "")

	parts := strings.Split(s, "/")
	return strings.Join(parts, "/")
}

func userExists(r *http.Request) bool {
	_, ok := r.Context().Value(keyUserName).(string)
	return ok
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

func withTransactionID(r *http.Request, transactionID uuid.UUID) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, keyTransactionID, transactionID)
	return r.WithContext(ctx)
}

func withUser(r *http.Request, u cabby.User) *http.Request {
	ctx := context.WithValue(context.Background(), keyUserName, u.Email)
	ctx = context.WithValue(ctx, keyCanAdmin, u.CanAdmin)
	ctx = context.WithValue(ctx, keyCollectionAccessList, u.CollectionAccessList)
	return r.WithContext(ctx)
}

// func validateUser(ts taxiiStorer, u, p string) (taxiiUser, bool) {
// 	tu, err := newTaxiiUser(ts, u, p)
// 	if err != nil {
// 		log.Error(err)
// 		return tu, false
// 	}
//
// 	if !tu.valid() {
// 		log.WithFields(log.Fields{"user": u}).Error("Invalid user")
// 		return tu, false
// 	}
//
// 	return tu, true
// }
