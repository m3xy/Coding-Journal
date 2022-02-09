package main

import "github.com/golang-jwt/jwt"

// API Response types

// Standard response issued on empty content requests - e.g. errors.
type StandardResponse struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

// Response used by authentication log in.
type AuthLogInResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	RedirectUrl  string `json:"redirect_url"`
	Expires      int64  `json:"expires"`
}

// JWT Claim types

type JwtClaims struct {
	ID    string `json:"userId"`
	Scope string
	jwt.StandardClaims
}
