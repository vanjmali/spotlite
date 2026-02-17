package middlewares

import (
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/account"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/respond"
)

// ValidatePermission represents a function which returns a middleware used to check if users have the permission,
// to do an action.
func ValidatePermission(allowedRoles ...account.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(roleIdKey).(account.Role)

			if !ok {
				logging.Securityf(r.Context(), "rbac_missing_role_in_context method=%s path=%s remote_addr=%s", r.Method, r.URL.Path, r.RemoteAddr)
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
				logging.Securityf(r.Context(), "rbac_access_denied method=%s path=%s role=%s remote_addr=%s", r.Method, r.URL.Path, userRole, r.RemoteAddr)
				_ = respond.Forbidden(w)
				return
			}

			if userRole == account.RoleAdmin {
				logging.Auditf(r.Context(), "admin_activity method=%s path=%s remote_addr=%s", r.Method, r.URL.Path, r.RemoteAddr)
			}

			next.ServeHTTP(w, r)
		})
	}
}
