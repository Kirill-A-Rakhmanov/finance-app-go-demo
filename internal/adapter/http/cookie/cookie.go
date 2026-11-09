package cookie

import (
	"net/http"
	"time"
)

const (
	AccessTokenName  = "access_token"
	RefreshTokenName = "refresh_token"
)

type Config struct {
	Secure          bool
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func (c Config) sameSite() http.SameSite {
	if c.Secure {
		return http.SameSiteStrictMode
	}

	return http.SameSiteLaxMode
}

func (c Config) SetTokensToCookie(w http.ResponseWriter, accessToken string, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: c.sameSite(),
		MaxAge:   int(c.AccessTokenTTL.Seconds()),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: c.sameSite(),
		MaxAge:   int(c.RefreshTokenTTL.Seconds()),
	})
}

func (c Config) ClearCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: c.sameSite(),
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: c.sameSite(),
		MaxAge:   -1,
	})
}
