package csrf

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
)

const (
	tokenLength = 32
)

type Token struct {
	Value     string
	ExpiresAt time.Time
}

func Generate() (string, error) {
	token := make([]byte, tokenLength)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func SetCookie(w http.ResponseWriter, token string, cfg config.CSRFConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.CookieName,
		Value:    hashToken(token),
		Path:     "/",
		MaxAge:   int(cfg.CookieTime.Seconds()),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func GetFromCookie(r *http.Request, cfg config.CSRFConfig) (string, error) {
	cookie, err := r.Cookie(cfg.CookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func Validate(requestToken, cookieToken string) bool {
	if requestToken == "" || cookieToken == "" {
		return false
	}
	return hashToken(requestToken) == cookieToken
}
