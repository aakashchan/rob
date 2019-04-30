package httperr

import (
	"encoding/json"
	"fmt"
	"net/http"
	c "rob/lib/common/constants"

	log "github.com/sirupsen/logrus"
)

type HttpError struct {
	Status     int
	Message    string
	DebugError string `json:",omitempty"`
}

func DB(w http.ResponseWriter, message string, err *error) {
	E(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %s", message), err)
}

func E(w http.ResponseWriter, status int, message string, err *error) {
	w.WriteHeader(status)
	he := HttpError{
		Status:  status,
		Message: message,
	}
	if c.DE {
		if err == nil {
			he.DebugError = "nil"
		} else {
			he.DebugError = (*err).Error()
		}
	}
	log.WithFields(log.Fields{
		"status":  he.Status,
		"message": he.Message,
		"error":   he.DebugError,
	}).Debug("Http error")

	data, er := json.Marshal(he)
	if er != nil {
		// This is the worst case. Writing error caused error
		// In this case simply dump error to body
		w.Write([]byte(fmt.Sprintf("Status=%d; Message=%s;", status, message)))
	} else {
		w.Write(data)
	}
}
