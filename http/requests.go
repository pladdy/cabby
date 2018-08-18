package http

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	cabby "github.com/pladdy/cabby2"
)

// per context docuentation, use a key type for context keys
type key int

const (
	userName           key = 0
	userCollectionList key = 1
	requestRange       key = 2
	canAdmin           key = 3
	jsonContentType        = "application/json"
	sixMonthsOfSeconds     = "63072000"
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

// func takeAddedAfter(r *http.Request) string {
// 	af := r.URL.Query()["added_after"]
//
// 	if len(af) > 0 {
// 		t, err := time.Parse(time.RFC3339Nano, af[0])
// 		if err == nil {
// 			return t.Format(time.RFC3339Nano)
// 		}
// 	}
// 	return ""
// }

// func takeCanAdmin(r *http.Request) bool {
// 	ca, ok := r.Context().Value(canAdmin).(bool)
// 	if !ok {
// 		return false
// 	}
// 	return ca
// }

// func takeCollectionAccess(r *http.Request) taxiiCollectionAccess {
// 	// get collection access map from context
// 	ca, ok := r.Context().Value(userCollectionList).(map[taxiiID]taxiiCollectionAccess)
// 	if !ok {
// 		return taxiiCollectionAccess{}
// 	}
//
// 	tid, err := taxiiIDFromString(takeCollectionID(r))
// 	if err != nil {
// 		return taxiiCollectionAccess{}
// 	}
// 	return ca[tid]
// }

func takeAPIRoot(r *http.Request) string {
	var apiRootIndex = 1
	return getToken(r.URL.Path, apiRootIndex)
}

func takeCollectionID(r *http.Request) string {
	var collectionIndex = 3
	return getToken(r.URL.Path, collectionIndex)
}

func takeObjectID(r *http.Request) string {
	var objectIDIndex = 5
	return getToken(r.URL.Path, objectIDIndex)
}

// func takeRequestRange(r *http.Request) taxiiRange {
// 	ctx := r.Context()
//
// 	tr, ok := ctx.Value(requestRange).(taxiiRange)
// 	if !ok {
// 		return taxiiRange{}
// 	}
// 	return tr
// }

// func takeStatusID(r *http.Request) string {
// 	var statusIndex = 3
// 	return getToken(r.URL.Path, statusIndex)
// }
//
// func takeStixID(r *http.Request) string {
// 	si := r.URL.Query()["match[id]"]
//
// 	if len(si) > 0 {
// 		return si[0]
// 	}
// 	return ""
// }
//
// func takeStixTypes(r *http.Request) []string {
// 	st := r.URL.Query()["match[type]"]
//
// 	if len(st) > 0 {
// 		return strings.Split(st[0], ",")
// 	}
// 	return []string{}
// }
//
func takeUser(r *http.Request) string {
	user, ok := r.Context().Value(userName).(string)
	if !ok {
		return ""
	}
	return user
}

//
// func takeVersion(r *http.Request) string {
// 	v := r.URL.Query()["match[version]"]
//
// 	if len(v) > 0 {
// 		return v[0]
// 	}
// 	return ""
// }

func trimSlashes(s string) string {
	re := regexp.MustCompile("^/")
	s = re.ReplaceAllString(s, "")

	re = regexp.MustCompile("/$")
	s = re.ReplaceAllString(s, "")

	parts := strings.Split(s, "/")
	return strings.Join(parts, "/")
}

func userExists(r *http.Request) bool {
	_, ok := r.Context().Value(userName).(string)
	if !ok {
		return false
	}
	return true
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

// func withTaxiiRange(r *http.Request, tr taxiiRange) *http.Request {
// 	ctx := context.WithValue(r.Context(), requestRange, tr)
// 	return r.WithContext(ctx)
// }

func withUser(r *http.Request, u cabby.User) *http.Request {
	ctx := context.WithValue(context.Background(), userName, u.Email)
	ctx = context.WithValue(ctx, canAdmin, u.CanAdmin)
	//ctx = context.WithValue(ctx, userCollectionList, tu.CollectionAccessList)
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
