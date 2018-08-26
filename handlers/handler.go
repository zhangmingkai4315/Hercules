package handlers

import (
	"fmt"
	"net/http"

	"github.com/zhangmingkai4315/hercules/utils"
)

// HealthCheckHandler for check health only
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"alive": true}`)
}

// RequestProxy will parse header and send the request based
// header info
func RequestProxy(w http.ResponseWriter, r *http.Request) {
	resp, err := utils.MakePrometheusRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, resp)
}
