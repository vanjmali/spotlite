package routers

import (
	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/user-service/handlers"
)

func HandleRequests(h *handlers.UserHandler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/register", h.HandleRegistration).Methods("POST")

	// the verify endpoint is defined as a get so it can redirect when link click happens,
	r.HandleFunc("/verify", h.HandleAccountVerification).Methods("GET", "POST")
	return r
}
