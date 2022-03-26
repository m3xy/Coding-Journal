package main

// ----------
// Authentication/User Endpoints
// ----------

// POST /auth/login response.
type AuthLogInResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	RedirectUrl  string `json:"redirect_url"`
	Expires      int64  `json:"expires"`
}

// GET /user/query
type QueryUsersResponse struct {
	StandardResponse
	Users []GlobalUser `json:"users"`
}

// ----------
// Submissions Endpoints
// ----------

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

// ----------
// Files Endpoints
// ----------

// GET /file/{id} body
type GetFileResponse struct {
	StandardResponse
	File *File `json:"file"`
}

// ----------
// Comments Endpoints
// ----------

// POST /file/{id}/newcomment body. {id} in the URL is the file id
type NewCommentResponse struct {
	StandardResponse
	ID uint `json:"id"`
}

// ----------
// Journal Endpoints
// ----------

// POST /journal/login response.
type JournalLogInResponse struct {
	ID string `json:"userId"`
}

// ----------
// Misc
// ----------

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
