package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

const (
	sessionCookieName = "quakemeup_session"
	sessionDuration   = 24 * time.Hour
)

// generateSessionID creates a new random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// getUserID gets the user ID from the session cookie
func getUserID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			id, err := generateSessionID()
			if err != nil {
				return "", err
			}
			return id, nil
		}
		return "", err
	}
	return cookie.Value, nil
}

// setSessionCookie sets the session cookie
func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(sessionDuration),
	})
}
