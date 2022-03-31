package main

// ----------
// Authentication/User Endpoints
// ----------

// POST /auth/register
type SignUpPostBody struct {
	Email        string `json:"email" validate:"email,required"`
	Password     string `json:"password,omitempty" validate:"min=8,max=64,ispw,required"`
	FirstName    string `json:"firstName" validate:"required,max=32"`
	LastName     string `json:"lastName" validate:"required,max=32"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	Organization string `json:"organization,omitempty"`
}

// POST /user/{id}/edit body.
type EditUserPostBody struct {
	Password     string `json:"password,omitempty" validate:"min=8,max=64,ispw"`
	FirstName    string `json:"firstName,omitempty" validate:"max=32"`
	LastName     string `json:"lastName,omitempty" validate:"max=32"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	Organization string `json:"organization,omitempty"`
}

// POST /auth/login body.
type AuthLoginPostBody struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	GroupNumber int    `json:"groupNumber,string"`
}

// POST /user/{id}/changepermissions
type ChangePermissionsPostBody struct {
	Permissions int `json:"permissions" validate:"min=0,max=4"`
}

// ----------
// Submissions Endpoints
// ----------

// POST /submissions/upload/zip
type UploadSubmissionByZipBody struct {
	Name           string `json:"name" validate:"required"`
	License        string `json:"license"`
	Abstract       string `json:"abstract"`
	ZipBase64Value string `json:"base64" validate:"base64,required"`

	Runnable         bool `json:"runnable"`
	TakesStdIn       bool `json:"takesStdIn"`
	TakesCmdLn       bool `json:"takseCmdLn"`
	TakesInputFile   bool `json:"takesInputFile"`
	ReqNetworkAccess bool `json:"reqNetworkAccess"`

	Tags      []string `json:"tags"`
	Authors   []string `json:"authors" validate:"required"`
	Reviewers []string `json:"reviewers"`
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
	Approved    bool   `json:"approved"`
	Base64Value string `json:"base64Value" validate:"required"`
}

// POST /submissions/{id}/approve
type UpdateSubmissionStatusBody struct {
	Status bool `json:"status"`
}

// ----------
// Comments Endpoints
// ----------

// POST /file/{id}/comment body. {id} in the URL is the file id
type NewCommentPostBody struct {
	ParentID    *uint  `json:"parentId,omitempty"` // optionally set for replies
	StartLine   int    `json:"startLine" validate:"min=0"`
	EndLine     int    `json:"endLine" validate:"min=0"`
	Base64Value string `json:"base64Value" validate:"required"`
}

// POST /file/{id}/comment/{commentId}/edit body
type EditCommentPostBody struct {
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
