package middleware

import (
	"net/http"
	c "rob/lib/common/constants"
	"rob/lib/common/httperr"
	"rob/lib/session"

	log "github.com/sirupsen/logrus"
)

func Auth(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var funcName = "middleware/authentication.go:Auth"
		log.Debugf("Enter: %s", funcName)
		defer log.Debugf("Exit: %s", funcName)

		sess := session.Instance(r)

		if sess.Values[c.Id] == nil {
			httperr.E(w, http.StatusUnauthorized, "User not logged in", nil)
			return
		}

		f.ServeHTTP(w, r)
	})
}

func NoAuth(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var funcName = "middleware/authentication.go:NoAuth"
		log.Debugf("Enter: %s", funcName)
		defer log.Debugf("Exit: %s", funcName)

		sess := session.Instance(r)

		if sess.Values[c.Id] != nil {
			httperr.E(w, http.StatusUnauthorized, "User Already logged in", nil)
			return
		}

		f.ServeHTTP(w, r)
	})
}
