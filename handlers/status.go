package handlers

import (
	"fmt"
	"net/http"
)

// HealthCheckHandler for check only
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"alive": true}`)
}
