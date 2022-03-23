package main

import "github.com/golang-jwt/jwt"

// --- API Responses --- //

// Standard response in content requests - e.g. errors.
type StandardResponse struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

type FormResponse struct {
	StandardResponse
	Fields []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"fields"`
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
	StandardResponse
	ID uint `json:"id"`
}

// GET /submissions/tags
type GetAvailableTagsResponse struct {
	StandardResponse
	Tags []string `json:"tags"`
}

// GET /submissions/query
type QuerySubmissionsResponse struct {
	StandardResponse
	Submissions []Submission `json:"submissions"` // submissions only contain ID and name
}

// POST /submissions/create
type UploadSubmissionResponse struct {
	StandardResponse
	SubmissionID uint `json:"ID"`
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
	LineNumber  int    `json:"lineNumber" validate:"min=0"`
	Base64Value string `json:"base64Value" validate:"required"`
}

// POST /submissions/create body
type UploadSubmissionBody struct {
	Name      string   `json:"name" validate:"required"`
	License   string   `json:"license"`
	Abstract  string   `json:"abstract"`
	Tags      []string `json:"tags"`
	Authors   []string `json:"authors" validate:"required"`
	Reviewers []string `json:"reviewers"`
	Files     []File   `json:"files"`
}

// POST /submissions/{id}/assignreviewers
type AssignReviewersBody struct {
	Reviewers []string `json:"reviewers" validate:"min=1"`
}

// POST /submissions/{id}/review
type UploadReviewBody struct {
	Approved    bool   `json:"approved" validate:"required"`
	Base64Value string `json:"base64Value" validate:"required"`
}

// POST /submissions/{id}/approve
type UpdateSubmissionStatusBody struct {
	Status bool `json:"status" validate:"required"`
}

// POST /submissions/upload/zip
type UploadSubmissionByZipBody struct {
	Name           string   `json:"name" validate:"required"`
	License        string   `json:"license"`
	Abstract       string   `json:"abstract"`
	Tags           []string `json:"tags"`
	Authors        []string `json:"authors" validate:"required"`
	Reviewers      []string `json:"reviewers"`
	ZipBase64Value string   `json:"base64" validate:"base64,required"`
}

// --- JWT Claim types --- //
type JwtClaims struct {
	ID       string `json:"userId"`
	UserType int    `json:"userType" validate:"min=0,max=4"`
	Scope    string
	jwt.StandardClaims
}
