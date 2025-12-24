package middlewares

import (
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/account"
)

// RequireAuthenticated is a middleware that ensures the user is authenticated (either admin or member).
func RequireAuthenticated(next http.HandlerFunc) http.Handler {
	return ValidateJWT(ValidatePermission(account.RoleAdmin, account.RoleMember)(next))
}

// RequireAdmin is a middleware that ensures the user has admin role.
func RequireAdmin(next http.HandlerFunc) http.Handler {
	return ValidateJWT(ValidatePermission(account.RoleAdmin)(next))
}
