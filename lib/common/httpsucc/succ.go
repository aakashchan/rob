package httpsucc

import (
	"fmt"
	"net/http"
)

func SuccWithMessage(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Status 200; Message %s;", message)))
}
