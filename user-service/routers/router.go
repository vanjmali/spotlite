package routers

import (
	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/user-service/handlers"
)

func HandleRequests(h *handlers.UserHandler, rth *handlers.RefreshTokenHandler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/register", h.HandleRegistration).Methods("POST")

	r.HandleFunc("/login", h.HandleLogin).Methods("POST")
	r.HandleFunc("/login/verify-otp", h.HandleVerifyLoginOtp).Methods("POST")
	r.HandleFunc("/refresh-token", rth.HandleRefreshToken).Methods("POST")

	// the verify endpoint is defined as a get so it can redirect when link click happens,
	r.HandleFunc("/verify", h.HandleAccountVerification).Methods("GET", "POST")

	r.HandleFunc("/change-password", h.HandleChangePassword).Methods("POST")
	return r
}
