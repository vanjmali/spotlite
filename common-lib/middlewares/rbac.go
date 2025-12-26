package middlewares

import (
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/account"
	"github.com/vanjmali/spotlite/common-lib/respond"
)

// ValidatePermission represents a function which returns a middleware used to check if users have the permission,
// to do an action.
func ValidatePermission(allowedRoles ...account.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(roleIdKey).(account.Role)

			if !ok {
				_ = respond.Unauthorized(w)
				return
			}

			isAllowed := false
			for _, role := range allowedRoles {
				if role == userRole {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				_ = respond.Forbidden(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
