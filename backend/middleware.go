package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

const (
	SECURITY_TOKEN_KEY = "X-FOREIGNJOURNAL-SECURITY-TOKEN"
)

// request context object for logged in users
type RequestContext struct {
	ID       string `validate:"required"`
	UserType int    `validate:"oneof=0 1 2 3 4"` // matches the 5 user types
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
		if ok, id, userType := validateWebToken(r.Header.Get("BearerToken"), CLAIM_BEARER); !ok {

			next.ServeHTTP(w, r)
		} else {
			if !isUnique(gormDb, &GlobalUser{}, "id", id) {
				ctx := context.WithValue(r.Context(), "data", &RequestContext{
					ID:       id,
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

// Struct and implementation for logger's response writer.
type StatusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *StatusResponseWriter) WriteHeader(statusCode int) {
	sw.statusCode = statusCode
	sw.ResponseWriter.WriteHeader(statusCode)
}

// Logged for incoming requests.
// Format: user user_type [method] [time] [code] host path query
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		user, usertype := "-", "-"
		if ctx, ok := r.Context().Value("data").(*RequestContext); ok && validate.Struct(ctx) == nil {
			user = ctx.ID
			userMap := map[int]string{0: "user", 1: "publisher", 2: "reviewer",
				3: "reviewer-publisher", 4: "editor"}
			usertype = userMap[ctx.UserType]
		}

		// Create response writer with given status.
		sw := &StatusResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Log the request's final result.
		defer func() {
			log.Printf("%s %v [%s] [%v] [%d] %s %s %s",
				user, usertype,
				r.Method, time.Since(start), sw.statusCode,
				r.Host, r.URL.Path, r.URL.RawQuery,
			)
		}()
		next.ServeHTTP(sw, r)
	})
}
