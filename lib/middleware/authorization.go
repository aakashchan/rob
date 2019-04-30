package middleware

import (
	"fmt"
	"net/http"
	c "rob/lib/common/constants"
	"rob/lib/common/httperr"
	"rob/lib/session"

	log "github.com/sirupsen/logrus"
)

func CheckAccess(allowedRoles ...int) (mw func(http.Handler) http.Handler) {
	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var funcName = "middleware/authorization.go:CheckAccess"
			log.WithFields(log.Fields{
				"allowedRoles": allowedRoles,
			}).Debugf("Enter: %s", funcName)
			defer log.Debugf("Exit: %s", funcName)

			allowed := false

			sess := session.Instance(r)
			roleId := sess.Values[c.RoleId].(int)
			for _, allowedRole := range allowedRoles {

				if roleId == allowedRole {
					allowed = true
					break
				}
			}

			if allowed != true {
				httperr.E(w, http.StatusUnauthorized, fmt.Sprintf("Access Denied for role %d", roleId), nil)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
	return
}
