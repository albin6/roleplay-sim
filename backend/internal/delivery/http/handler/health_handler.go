package handler

import "net/http"

// HealthHandler returns 200 OK for readiness/liveness probes.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
