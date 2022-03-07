package main

import (
	"context"
	"net/http"
)

const (
	SECURITY_TOKEN_KEY = "X-FOREIGNJOURNAL-SECURITY-TOKEN"
)

// request context object for logged in users
type RequestContext struct {
	ID string `validate:"required"`
	UserType int `validate:"required"`
}

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
		if ok, id, userType := validateWebToken(r.Header.Get("Authorization"), CLAIM_BEARER); !ok {
			next.ServeHTTP(w, r)
		} else {
			if !isUnique(gormDb, &GlobalUser{}, "id", id) {
				ctx := context.WithValue(r.Context(), "data", RequestContext{
					ID: id,
					UserType: userType,
				})
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		}
		return
	})
}
