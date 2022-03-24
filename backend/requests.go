package main

// ----------
// Authentication/User Endpoints
// ----------

// POST /auth/login body.
type AuthLoginPostBody struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	GroupNumber int    `json:"groupNumber,string"`
}

// ----------
// Submissions Endpoints
// ----------

// POST /submissions/upload/zip
type UploadSubmissionByZipBody struct {
	Name           string   `json:"name" validate:"required"`
	License        string   `json:"license"`
	Abstract       string   `json:"abstract"`
	Tags           []string `json:"tags"`
	Authors        []string `json:"authors" validate:"required"`
	Reviewers      []string `json:"reviewers"`
	ZipBase64Value string   `json:"base64" validate:"base64,required"`
	Runnable       bool     `json:"runnable"`
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
	Runnable  bool     `json:"runnable"`
}

// ----------
// Approval Endpoints
// ----------

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

// ----------
// Comments Endpoints
// ----------

// POST /file/{id}/newcomment body. {id} in the URL is the file id
type NewCommentPostBody struct {
	ParentID    *uint  `json:"parentId,omitempty"` // optionally set for replies
	LineNumber  int    `json:"lineNumber" validate:"min=0"`
	Base64Value string `json:"base64Value" validate:"required"`
}

// ----------
// Journal Endpoints
// ----------

// POST /journal/login body.
type JournalLoginPostBody struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	GroupNumber string `json:"groupNumber"`
}
