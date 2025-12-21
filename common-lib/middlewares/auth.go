package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/user-service/entities"
)

// ctxKey type is used to add data in the context while avoiding conflicts with other services
// that rely on the context as well.
type ctxKey string

const (
	StatusKey ctxKey = "status"
	UserIdKey ctxKey = "userId"
	RoleIdKey ctxKey = "role"
)

func ValidateJWT(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pubKey, _ := utils.GetPublicKey()

		tStr := extractToken(r.Header.Get("Authorization"))
		t, err := jwt.Parse(tStr, func(token *jwt.Token) (interface{}, error) {
			// final check to avoid "JWT Algorithm Confusion Attack"
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return pubKey, nil
		})

		if err != nil || !t.Valid {
			_ = respond.Unauthorized(w)
			return
		}

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			_ = respond.Unauthorized(w)
			return
		}

		status, ok := claims["status"].(string)
		if !ok {
			_ = respond.Unauthorized(w)
			return
		}
		ctx := context.WithValue(r.Context(), StatusKey, status)

		userId, ok := claims["sub"].(string)
		if !ok {
			_ = respond.Unauthorized(w)
			return
		}
		ctx = context.WithValue(ctx, UserIdKey, userId)

		roleStr, ok := claims["role"].(string)
		if !ok {
			_ = respond.Unauthorized(w)
			return
		}
		userRole := entities.UserRole(roleStr)
		ctx = context.WithValue(ctx, RoleIdKey, userRole)
		r = r.WithContext(ctx)

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
