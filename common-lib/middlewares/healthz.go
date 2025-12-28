package middlewares

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/respond"
)

// HandleHealthz sets up a router with a /healthz endpoint for health checks.
// It is for used internal monitoring purposes.
func HandleHealthz(r *mux.Router) http.Handler {
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_ = respond.OkJson(w, map[string]string{"status": "ok"})
	}).Methods("GET")
	return r
}
