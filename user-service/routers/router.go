package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/user-service/handlers"
)

// HandleRequests wires HTTP routes to user handlers.
func HandleRequests(h *handlers.UserHandler, rth *handlers.RefreshTokenHandler, prh *handlers.PasswordRecoveryHandler) http.Handler {
	r := mux.NewRouter()
	middlewares.HandleHealthz(r)

	// create a subrouter for API routes to attach telemetry and rate limiting
	api := r.PathPrefix("/").Subrouter()

	telemetry.AttachMuxTracing(api, "user-service")

	api.HandleFunc("/register", h.HandleRegistration).Methods("POST")

	api.HandleFunc("/login", h.HandleLogin).Methods("POST")
	api.HandleFunc("/login/verify-otp", h.HandleVerifyLoginOtp).Methods("POST")
	api.HandleFunc("/login/resend-otp", h.HandleResendOtp).Methods("POST")

	api.HandleFunc("/check-email", h.HandleCheckEmail).Methods("GET")

	api.HandleFunc("/password-recovery/request", prh.HandleRequestPasswordReset).Methods("POST")
	api.HandleFunc("/password-recovery/validate", prh.HandleValidateRecoveryToken).Methods("POST")
	api.HandleFunc("/password-recovery/reset", prh.HandleResetPassword).Methods("POST")

	api.HandleFunc("/refresh-token", rth.HandleRefreshToken).Methods("POST")
	api.HandleFunc("/logout", h.HandleLogout).Methods("POST")

	// the verify endpoint is defined as a get so it can redirect when link click happens.
	api.HandleFunc("/verify", h.HandleAccountVerification).Methods("GET", "POST")

	api.Handle("/change-password", middlewares.RequireAuthenticated(h.HandleChangePassword)).Methods("PATCH")

	return r
}
