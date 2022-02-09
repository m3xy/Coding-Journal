package main

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	SECURITY_TOKEN_KEY = "X-FOREIGNJOURNAL-SECURITY-TOKEN"
)

// Middleware for user authentication and security key verification.
func journalMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validateSecurityKey(gormDb, r.Header.Get(SECURITY_TOKEN_KEY)) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// Middleware for strict access token validation.
func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok, id := validateWebToken(r.Header.Get("Authorization"), CLAIM_BEARER); !ok {
			goto FAILURE
		} else {
			if !isUnique(gormDb, &GlobalUser{}, "id", id) {
				ctx := context.WithValue(r.Context(), "userId", id)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				goto FAILURE
			}
		}

		return
	FAILURE:
		_ = json.NewEncoder(w).Encode(StandardResponse{
			Message: "User is not logged in!",
			Error:   true,
		})
		w.WriteHeader(http.StatusUnauthorized)
	})
}
