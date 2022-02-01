package main

import (
	"context"
	"net/http"
)

const (
	SECURITY_TOKEN_KEY = "X-FOREIGNJOURNAL-SECURITY-TOKEN"
)

// Middleware for user authentication and security key verification.
func authenticationMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			if r.Header.Get("user") != "" {
				ctx := context.WithValue(r.Context(), "user", r.Header.Get("user"))
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		}
	})
}
