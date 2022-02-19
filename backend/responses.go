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

// POST /file/{id}/newcomment body. {id} in the URL is the file id
type NewCommentResponse struct {
	ID uint `json:"id"`
}

// --- Request bodies --- //

// POST /auth/login body.
type AuthLoginPostBody struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	GroupNumber int    `json:"groupNumber,string"`
}

// POST /journal/login body.
type JournalLoginPostBody struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	GroupNumber string `json:"groupNumber"`
}

// POST /file/{id}/newcomment body. {id} in the URL is the file id
type NewCommentPostBody struct {
	ParentID    *uint  `json:"parentId,omitempty"` // optionally set for replies
	Base64Value string `json:"base64Value"`
}

// --- JWT Claim types --- //
type JwtClaims struct {
	ID    string `json:"userId"`
	Scope string
	jwt.StandardClaims
}
