package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var (
	store  *sessions.CookieStore
	name   = "sesid"
	secret = "@r4B?EhaSEh_drudR7P_hdsa=s#s2Pah"
	esec   = "@71S_D_@3d86!@0-"
)

func init() {
	store = sessions.NewCookieStore([]byte(secret), []byte(esec))
	store.Options.HttpOnly = true
	store.Options.MaxAge = 0
}

func Instance(r *http.Request) *sessions.Session {
	session, _ := store.Get(r, name)

	return session
}

func Empty(sess *sessions.Session) {
	// Clear out all stored values in the cookie
	for k := range sess.Values {
		delete(sess.Values, k)
	}
}
