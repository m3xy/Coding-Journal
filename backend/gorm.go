package main

import (
	"log"
	"time"

	"fmt"
	"os"

	uuid "github.com/satori/go.uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var gormDb *gorm.DB

type User struct {
	ID           string `gorm:"type:varchar(191);not null;primaryKey" json:"userId"`
	Email        string `gorm:"uniqueIndex;unique;not null" json:"email"`
	Password     string `gorm:"not null" json:"password" validate:"min=8,max=64,validpw"`
	FirstName    string `validate:"nonzero,max=32" json:"firstname"`
	LastName     string `validate:"nonzero,max=32" json:"lastname"`
	UserType     int    `json:"usertype"`
	PhoneNumber  string `json:"phonenumber"`
	Organization string `json:"organization"`

	CreatedAt time.Time      `json:",omitempty"`
	UpdatedAt time.Time      `json:",omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:",omitempty"`
}

type Server struct {
	GroupNumber int    `gorm:"not null;primaryKey"`
	Token       string `gorm:"size:1028;not null"`
	Url         string `gorm:"not null; size:512"`

	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type GlobalID struct {
	ID     string `gorm:"not null;primaryKey" json:"globalId"`
	UserID string `json:"userId"`
	User   User

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type QueryUser struct {
	ID       string
	Email    string
	Password string
}

// Structure for code Submissions
type Submission struct {
	gorm.Model
	// name of the submission
	Name string `gorm:"not null;size:128;index" json:"submissionName"`
	// license which the code is published under
	License string `gorm:"size:64" json:"license"`
	// an array of the submission's files
	// File      []File     `gorm:"foreignKey:SubmissionID;references:ID"`
	// an array of the submission's file paths
	FilePaths []string `json:"filePaths"`
	// an array of the submissions's authors
	AuthorIDs   []string `gorm:"-" json:"AuthorIDs"`
	Authors    []User     `gorm:"many2many:authors_submission"`
	// an array of the submission's reviewers
	ReviewerIDs   []string `gorm:"-" json:"ReviewerIDs"`
	Reviewers  []User     `gorm:"many2many:reviewers_submission"`
	// tags for organizing/grouping code submissions
	Categories []Category `gorm:"foreignKey:SubmissionID; references:ID" json:"categories"`
	// metadata about the submission
	MetaData *SubmissionData `gorm:"-" json:"metadata"`
}

// Supergroup compliant code submissions (never stored in db)
type SupergroupSubmission struct {
	// name of the submission
	Name string `json:"name"`
	// metadata about the submission
	MetaData *SupergroupSubmissionData `json:"metadata"`
	// file objects which are members of the submission
	Files []*SupergroupFile `json:"files"`
}

// structure for meta-data of the submission. matches the structure of the submission's
// JSON data file. This struct is never stored in the db
type SubmissionData struct {
	// abstract for the submission, to be displayed upon opening of any given submission
	Abstract string `json:"abstract"`
	// reviewer comments on the overall submission
	Reviews []*Comment `json:"reviews"`
}

// supergroup compliant structure for meta-data of the submission
type SupergroupSubmissionData struct {
	// date the code submission was created
	CreationDate string `json:"creationDate"`
	// names of the authors listed on the submission
	AuthorNames []string `json:"authorNames"`
	// tags for organizing/grouping code submissions (does not access db, so doesnt use Category type)
	Categories []string `json:"categories"`
	// abstract for the submission, to be displayed upon opening of any given submission
	Abstract string `json:"abstract"`
	// license which the code is published under
	License string `json:"license"`
}

// Structure for code files
type File struct {
	gorm.Model
	// id of the submission this file is a part of
	SubmissionID int `json:"submissionId"`
	// name of the submission this file is a part of
	SubmissionName string `json:"submissionName"` // TODO: remove if obselete
	// relative path to the file from the root of the submission's file structure
	Path string `json:"filePath" db:"filePath"`
	// base name of the file with extension
	Name string `json:"filename"`
	// content of the file encoded as a Base64 string (non-db field)
	Base64Value string `gorm:"-" json:"base64Value"`
	// structure to hold the data from the file's metadata file
	MetaData *FileData `gorm:"-" json:"metadata"`
}

// Supergroup compliant file structure (never stored in db)
type SupergroupFile struct {
	// name of the file as a string
	Name string `json:"filename"`
	// file content as a base64 encoded string
	Base64Value string `json: "base64Value"`
}

// structure to hold json data from data files (never stored in db)
type FileData struct {
	// stores comments for the given code file
	Comments []*Comment `json:"comments"`
}

// Structure for user comments on code (not written to db)
type Comment struct {
	// author of the comment as an id
	AuthorId string `json:"author"`
	// time that the comment was recorded as a string
	Time string `json:"time"`
	// content of the comment as a string
	Base64Value string `json:"base64Value"`
	// replies TEMP: maybe don't allow nested replies?
	Replies []*Comment `json:"replies"`
}

// stores submission tags (i.e. networking, java, python, etc.)
type Category struct {
	gorm.Model
	Tag string `gorm:"uniqueIndex;unique" json:"category"` // actual tag
	SubmissionID int
}

func gormInit(dbname string, logger logger.Interface) (*gorm.DB, error) {
	// Set MySQL info in DSN format according to Go MySQL Drive -
	// user:password@protocol(host:port)/dbname?[param1=val...]
	mysqlInfo := fmt.Sprintf("%s/%s?%s", os.Getenv("DATABASE_URL"), dbname,
		getDbParams(DB_PARAMS)) // Setting this to allow prepared statements.
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		goto ERR
	}
	err = db.AutoMigrate(&User{}, &Server{}, &GlobalID{})
	if err != nil {
		goto ERR
	}
	return db, nil

ERR:
	log.Fatalf("SQL initialization error: %v", err)
	return nil, err
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.NewV4().String()
	}
	return
}

func gormClear(db *gorm.DB) error {
	tables := []interface{}{&GlobalID{}, &User{}}
	for _, table := range tables {
		res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).
			Unscoped().Delete(table)
		if err := res.Error; err != nil {
			return err
		}
	}
	return nil
}

func isUnique(db *gorm.DB, table interface{}, varname string, val string) bool {
	var exists bool
	if err := db.Model(table).Select("count(*) > 0").Where(varname+" = ?", val).Find(&exists).Error; err != nil {
		return false
	} else {
		return !exists
	}
}
