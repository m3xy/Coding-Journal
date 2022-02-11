package main

import "github.com/golang-jwt/jwt"

// --- API Responses --- //

// Standard response in content requests - e.g. errors.
type StandardResponse struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

// POST /auth/login response.
type AuthLogInResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	RedirectUrl  string `json:"redirect_url"`
	Expires      int64  `json:"expires"`
}

// POST /journal/login response.
type JournalLogInResponse struct {
	ID string `json:"userId"`
}

// --- Request bodies --- //

// POST /auth/login body.
type AuthLoginPostBody struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	GroupNumber int    `json:"groupNumber"`
}

// POST /journal/login body.
type JournalLoginPostBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// --- JWT Claim types --- //
type JwtClaims struct {
	ID    string `json:"userId"`
	Scope string
	jwt.StandardClaims
}
