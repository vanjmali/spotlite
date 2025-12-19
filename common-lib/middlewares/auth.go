package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vanjmali/spotlite/common-lib/respond"
	utils2 "github.com/vanjmali/spotlite/common-lib/utils"
)

// ctxKey type is used to add data in the context while avoiding conflicts with other services
// that rely on the context as well.
type ctxKey string

const (
	StatusKey ctxKey = "status"
	UserIdKey ctxKey = "userId"
	RoleIdKey ctxKey = "role"
)

func ValidateJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pubKey, err := utils2.GetPublicKey()
		if err != nil {
			_ = respond.InternalServerError(w)
			fmt.Println(fmt.Sprintf("an error has occurred while fetching public key: %s", err))
			return
		}

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

		role, ok := claims["role"].(string)
		if !ok {
			_ = respond.Unauthorized(w)
			return
		}
		ctx = context.WithValue(ctx, RoleIdKey, role)
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
