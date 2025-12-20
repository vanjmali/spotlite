package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/handlers"
)

//nolint:unused // These are helper functions for future route protection
func requireAuthenticated(next http.HandlerFunc) http.Handler {
	return middlewares.ValidateJWT(middlewares.ValidatePermission(entities.RoleAdmin, entities.RoleMember)(next))
}

//nolint:unused // These are helper functions for future route protection
func requireAdmin(next http.HandlerFunc) http.Handler {
	return middlewares.ValidateJWT(middlewares.ValidatePermission(entities.RoleAdmin)(next))
}

// HandleRequests wires HTTP routes to user handlers.
func HandleRequests(h *handlers.UserHandler, rth *handlers.RefreshTokenHandler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/register", h.HandleRegistration).Methods("POST")

	r.HandleFunc("/login", h.HandleLogin).Methods("POST")
	r.HandleFunc("/login/verify-otp", h.HandleVerifyLoginOtp).Methods("POST")
	r.HandleFunc("/refresh-token", rth.HandleRefreshToken).Methods("POST")

	// the verify endpoint is defined as a get so it can redirect when link click happens,
	r.HandleFunc("/verify", h.HandleAccountVerification).Methods("GET", "POST")
	return r
}
