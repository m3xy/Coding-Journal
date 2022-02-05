package main

import (
	"log"
	"strconv"
	"time"

	"fmt"
	"os"

	uuid "github.com/satori/go.uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const (
	USERTYPE_NIL                = 0
	USERTYPE_PUBLISHER          = 1
	USERTYPE_REVIEWER           = 2
	USERTYPE_REVIEWER_PUBLISHER = 3
	USERTYPE_USER               = 4
)

var gormDb *gorm.DB

type User struct {
	ID           uint   `gorm:"primaryKey" json:"-"`
	GlobalUserID string `json:"-"`
	Email        string `gorm:"uniqueIndex;unique;not null" json:"Email"`
	Password     string `gorm:"not null" json:"Password,omitempty" validate:"min=8,max=64,validpw"`
	FirstName    string `validate:"nonzero,max=32" json:"FirstName"`
	LastName     string `validate:"nonzero,max=32" json:"LastName"`
	UserType     int    `gorm:"default:4" json:"UserType"`
	PhoneNumber  string `json:"PhoneNumber"`
	Organization string `json:"Organization"`

	CreatedAt time.Time      `json:",omitempty"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Server struct {
	GroupNumber int    `gorm:"not null;primaryKey"`
	Token       string `gorm:"size:1028;not null"`
	Url         string `gorm:"not null; size:512"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type GlobalUser struct {
	ID          string       `gorm:"not null;primaryKey;type:varchar(191)" json:"UserID"`
	FullName    string       `json:"FullName"`
	User        User         `json:"Profile"`
	Submissions []Submission `gorm:"many2many:authors_submission"`

	CreatedAt time.Time      `json:"CreatedAt"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Structure for code Submissions
type Submission struct {
	gorm.Model
	// name of the submission
	Name string `gorm:"not null;size:128;index" json:"submissionName"`
	// license which the code is published under
	License string `gorm:"size:64" json:"license"`
	// an array of the submission's files
	Files []File `gorm:"foreignKey:SubmissionID; references:ID"`
	// an array of the submissions's authors
	Authors []GlobalUser `gorm:"many2many:authors_submission"`
	// an array of the submission's reviewers
	Reviewers []GlobalUser `gorm:"many2many:reviewers_submission"`
	// tags for organizing/grouping code submissions
	Categories []string `gorm:"-" json:"categories"`
	// metadata about the submission
	MetaData *SubmissionData `gorm:"-" json:"metadata,omitempty"`
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

// struct for code files
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
	Base64Value string `json:"base64Value"`
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
	Tag          string `gorm:"column" json:"category"` // actual content of the tag
	SubmissionID uint   `gorm:"foreignKey:SubmissionID; references:Submissions.ID"`
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
	err = db.AutoMigrate(&GlobalUser{}, &User{}, &Server{}, &Submission{}, &Category{}, &File{})
	if err != nil {
		goto ERR
	}
	return db, nil

ERR:
	log.Fatalf("SQL initialization error: %v", err)
	return nil, err
}

func (u *GlobalUser) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = strconv.Itoa(TEAM_ID) + uuid.NewV4().String()
	}
	return
}

func gormClear(db *gorm.DB) error {
	// deletes submissions w/ Authors/Reviewers Associations
	var submissions []Submission
	if err := db.Find(&submissions).Error; err != nil {
		return err
	}
	for _, submission := range submissions {
		db.Select(clause.Associations).Delete(&submission)
	}
	// Deletes main tables
	tables := []interface{}{&File{}, &Category{}, &User{}, &GlobalUser{}}
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
