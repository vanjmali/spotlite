package routers

import (
	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/user-service/handlers"
)

func HandleRequests(h *handlers.UserHandler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/register", h.HandleRegistration).Methods("POST")
	r.HandleFunc("/login", h.HandleLogin).Methods("POST")
	return r
}
