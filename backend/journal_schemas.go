package main

import "time"

// Supergroup compliant user type
type SupergroupUser struct {
	ID           string `json:"id" gorm:"global_users.id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`

	Email        string `json:"email" gorm:"users.email"`
	PhoneNumber  string `json:"phoneNumber"`
	Organization string `json:"organization"`
}

// Supergroup compliant code submissions (never stored in db)
type SupergroupSubmission struct {
	Name         string                   `json:"name"`
	MetaData     SupergroupSubmissionData `json:"metadata"`
	CodeVersions []SupergroupCodeVersion  `json:"codeVersions"`
}

// supergroup compliant structure for meta-data of the submission
type SupergroupSubmissionData struct {
	CreationDate time.Time `json:"creationDate"`
	Abstract     string    `json:"abstract"`
	License      string    `json:"license"`

	Categories []string           `json:"categories"`
	Authors    []SuperGroupAuthor `json:"authors"`
}

type SuperGroupAuthor struct {
	ID      string `json:"userId"`
	Journal string `json:"journal"`
}

// struct to store a supergroup compliant version of a submission
type SupergroupCodeVersion struct {
	TimeStamp time.Time        `json:"timestamp"`
	Files     []SupergroupFile `json:"files"`
}

// Supergroup compliant file structure (never stored in db)
type SupergroupFile struct {
	Name        string `json:"filename"` // actually a file path not basename
	Base64Value string `json:"base64Value"`
}
