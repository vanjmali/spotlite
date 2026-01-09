package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vanjmali/spotlite/common-lib/utils"
)

func refreshCookieName() string {
	return utils.MustGetEnv("APP_REFRESH_COOKIE_NAME")
}

func refreshCookieDomain() string {
	return utils.GetEnv("APP_REFRESH_COOKIE_DOMAIN", "")
}

func refreshCookieSecure() bool {
	raw := utils.MustGetEnv("APP_REFRESH_COOKIE_SECURE")
	secure, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}
	return secure
}

func refreshCookieSameSite() http.SameSite {
	raw := strings.ToLower(strings.TrimSpace(utils.MustGetEnv("APP_REFRESH_COOKIE_SAMESITE")))
	switch raw {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func setRefreshCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	maxAge := max(int(time.Until(expiresAt).Seconds()), 0)

	c := &http.Cookie{
		Name:     refreshCookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   refreshCookieSecure(),
		SameSite: refreshCookieSameSite(),
		Expires:  expiresAt,
		MaxAge:   maxAge,
	}

	if domain := refreshCookieDomain(); domain != "" {
		c.Domain = domain
	}

	http.SetCookie(w, c)
}

func clearRefreshCookie(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:     refreshCookieName(),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   refreshCookieSecure(),
		SameSite: refreshCookieSameSite(),
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}

	if domain := refreshCookieDomain(); domain != "" {
		c.Domain = domain
	}

	http.SetCookie(w, c)
}
