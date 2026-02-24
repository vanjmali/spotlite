package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vanjmali/spotlite/common-lib/account"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/utils"
)

// ctxKey type is used to add data in the context while avoiding conflicts with other services
// that rely on the context as well.
type ctxKey string

const (
	statusKey   ctxKey = "status"
	userIdKey   ctxKey = "userId"
	roleIdKey   ctxKey = "role"
	usernameKey ctxKey = "username"
)

func ValidateJWT(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pubKey, _ := utils.GetPublicKey()

		tStr := extractToken(r.Header.Get("Authorization"))
		if tStr == "" {
			logging.Securityf(r.Context(), "auth_missing_token method=%s path=%s remote_addr=%s", r.Method, r.URL.Path, r.RemoteAddr)
			_ = respond.Unauthorized(w)
			return
		}

		t, err := jwt.Parse(tStr, func(token *jwt.Token) (interface{}, error) {
			// final check to avoid "JWT Algorithm Confusion Attack"
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return pubKey, nil
		})

		if err != nil || !t.Valid {
			logging.Securityf(r.Context(), "auth_invalid_or_expired_token method=%s path=%s remote_addr=%s", r.Method, r.URL.Path, r.RemoteAddr)
			_ = respond.Unauthorized(w)
			return
		}

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			logging.Securityf(r.Context(), "auth_invalid_token_claims method=%s path=%s remote_addr=%s", r.Method, r.URL.Path, r.RemoteAddr)
			_ = respond.Unauthorized(w)
			return
		}

		status, ok := claims["status"].(string)
		if !ok {
			logging.Securityf(r.Context(), "auth_missing_status_claim method=%s path=%s remote_addr=%s", r.Method, r.URL.Path, r.RemoteAddr)
			_ = respond.Unauthorized(w)
			return
		}
		ctx := context.WithValue(r.Context(), statusKey, status)

		userId, ok := claims["sub"].(string)
		if !ok {
			logging.Securityf(r.Context(), "auth_missing_subject_claim method=%s path=%s remote_addr=%s", r.Method, r.URL.Path, r.RemoteAddr)
			_ = respond.Unauthorized(w)
			return
		}
		ctx = context.WithValue(ctx, userIdKey, userId)

		roleStr, ok := claims["role"].(string)
		if !ok {
			logging.Securityf(r.Context(), "auth_missing_role_claim method=%s path=%s remote_addr=%s", r.Method, r.URL.Path, r.RemoteAddr)
			_ = respond.Unauthorized(w)
			return
		}
		userRole := account.Role(roleStr)
		ctx = context.WithValue(ctx, roleIdKey, userRole)

		username, ok := claims["username"].(string)
		if !ok {
			logging.Securityf(r.Context(), "auth_missing_username_claim method=%s path=%s remote_addr=%s", r.Method, r.URL.Path, r.RemoteAddr)
			_ = respond.Unauthorized(w)
			return
		}
		ctx = context.WithValue(ctx, usernameKey, username)

		r = r.WithContext(ctx)
		logging.Auditf(r.Context(), "auth_token_validated method=%s path=%s user_id=%s role=%s username=%s", r.Method, r.URL.Path, userId, userRole, username)

		next.ServeHTTP(w, r)
	}
}

func extractToken(authorizationHeader string) string {
	bearerToken := strings.Split(authorizationHeader, " ")

	if len(bearerToken) == 2 {
		return bearerToken[1]
	}

	return ""
}

// GetUsernameFromContext retrieves the user ID from the request context.
func GetUsernameFromContext(ctx context.Context) string {
	username, _ := ctx.Value(usernameKey).(string)
	return username
}

// GetUserIdFromContext retrieves the user ID from the request context.
func GetUserIdFromContext(ctx context.Context) string {
	userId, _ := ctx.Value(userIdKey).(string)
	return userId
}

// GetUserStatusFromContext retrieves the user account status from the request context.
func GetUserStatusFromContext(ctx context.Context) string {
	status, _ := ctx.Value(statusKey).(string)
	return status
}

// GetUserRoleFromContext retrieves the user role from the request context.
func GetUserRoleFromContext(ctx context.Context) account.Role {
	role, _ := ctx.Value(roleIdKey).(account.Role)
	return role
}

// ContextWithUserID is a helper function for testing that adds a user ID to a context.
// It should only be used in test code.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIdKey, userID)
}
