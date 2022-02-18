package main

import (
	"errors"
	"log"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fmt"
	"os"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	_ "github.com/go-sql-driver/mysql"
)

const (
	TEAM_ID                     = 11
	USERTYPE_NIL                = 0
	USERTYPE_PUBLISHER          = 1
	USERTYPE_REVIEWER           = 2
	USERTYPE_REVIEWER_PUBLISHER = 3
	USERTYPE_USER               = 4

	// Password related
	HASH_COST = 8
)

var gormDb *gorm.DB
var DB_PARAMS map[string]string = map[string]string{
	"interpolateParams": "true",
	"parseTime":         "true",
}

// User profile and personal information.
type User struct {
	ID           uint   `gorm:"primaryKey" json:"-"`
	GlobalUserID string `json:"-"`
	Email        string `gorm:"uniqueIndex;unique;not null" json:"email" validate:"isemail"`
	Password     string `gorm:"not null" json:"password,omitempty" validate:"min=8,max=64,ispw"`
	FirstName    string `validate:"nonzero,max=32" json:"firstName"`
	LastName     string `validate:"nonzero,max=32" json:"lastName"`
	UserType     int    `gorm:"default:4" json:"userType"`
	PhoneNumber  string `json:"phoneNumber"`
	Organization string `json:"organization"`

	CreatedAt time.Time      `json:",omitempty"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User global identification.
type GlobalUser struct {
	ID                  string       `gorm:"not null;primaryKey;type:varchar(191)" json:"userId"`
	FullName            string       `json:"fullName"`
	User                User         `json:"Profile,omitempty"`
	AuthoredSubmissions []Submission `gorm:"many2many:authors_submission" json:"-"`
	ReviewedSubmissions []Submission `gorm:"many2many:reviewers_submission" json:"-"`

	CreatedAt time.Time      `json:"CreatedAt"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Foreign journals which this journal can connect to.
type Server struct {
	GroupNumber int    `gorm:"not null;primaryKey"`
	Token       string `gorm:"size:1028;not null"`
	Url         string `gorm:"not null; size:512"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Structure for code Submissions
type Submission struct {
	gorm.Model
	// name of the submission
	Name string `gorm:"not null;size:128;index" json:name"`
	// license which the code is published under
	License string `gorm:"size:64" json:"license"`
	// an array of the submission's files
	Files []File `json:"files,omitempty"`
	// an array of the submissions's authors
	Authors []GlobalUser `gorm:"many2many:authors_submission" json:"authors,omitempty"`
	// an array of the submission's reviewers
	Reviewers []GlobalUser `gorm:"many2many:reviewers_submission" json:"reviewers,omitempty"`
	// tags for organizing/grouping code submissions
	Categories []string `gorm:"-" json:"categories,omitempty"`
	// metadata about the submission
	MetaData *SubmissionData `gorm:"-" json:"metaData,omitempty"`
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
	SubmissionID uint `json:"submissionId"`
	// relative path to the file from the root of the submission's file structure
	Path string `json:"path"`
	// base name of the file with extension
	Name string `json:"name"`
	// content of the file encoded as a Base64 string (non-db field)
	Base64Value string `gorm:"-" json:"base64Value"`
	// structure to hold the user comments on the file
	Comments []Comment `json:"comments"`
}

// Supergroup compliant file structure (never stored in db)
type SupergroupFile struct {
	// name of the file as a string
	Name string `json:"filename"`
	// file content as a base64 encoded string
	Base64Value string `json:"base64Value"`
}

// Structure for user comments on code (not written to db)
type Comment struct {
	gorm.Model
	// author of the comment as an id
	AuthorID string `json:"author"`
	// file which the comment belongs to
	FileID uint `json:"fileId"`
	// content of the comment as a string
	Base64Value string `gorm:"type:mediumtext" json:"base64Value"`
	ParentID *uint `gorm:"default:NULL"` // pointer so it can be nil
	Comments []Comment `gorm:"foreignKey:ParentID" json:"comments"`
}

// stores submission tags (i.e. networking, java, python, etc.)
// uniqueIndex:idx_first_second specifies the first and second column as a unique pair
type Category struct {
	Tag          string `gorm:"column;uniqueIndex:idx_first_second" json:"category"` // actual content of the tag
	SubmissionID uint   `gorm:"uniqueIndex:idx_first_second" json:"-"`
}

// ---- Database and reflect utilities ----

// Initialise database - open connection, migrate tables, set logger.
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
	err = db.AutoMigrate(&GlobalUser{}, &User{}, &Server{}, &Submission{}, &Category{}, &File{}, &Comment{})
	if err != nil {
		goto ERR
	}
	return db, nil

ERR:
	log.Fatalf("SQL initialization error: %v", err)
	return nil, err
}

// Pre-user creation hook - generate UUID with journal number appended.
func (u *GlobalUser) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = strconv.Itoa(TEAM_ID) + uuid.NewV4().String()
	}
	return
}

// Clear every table rows in the database.
func gormClear(db *gorm.DB) error {
	// deletes comments w/ associations 
	var comments []Comment
	if err := db.Find(&comments).Error; err != nil {
		return err
	}
	for _, comment := range comments {
		db.Select(clause.Associations).Unscoped().Delete(&comment)
	}
	// deletes files w/ associations 
	var files []File
	if err := db.Find(&files).Error; err != nil {
		return err
	}
	for _, file := range files {
		db.Select(clause.Associations).Unscoped().Delete(&file)
	}
	// deletes submissions w/ associations
	var submissions []Submission
	if err := db.Find(&submissions).Error; err != nil {
		return err
	}
	for _, submission := range submissions {
		db.Select(clause.Associations).Unscoped().Delete(&submission)
	}
	// Deletes main tables
	tables := []interface{}{&Comment{}, &File{}, &Category{}, &User{}, &GlobalUser{}, &Submission{}}
	for _, table := range tables {
		res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).
			Unscoped().Delete(table)
		if err := res.Error; err != nil {
			return err
		}
	}
	return nil
}

// Returns true if a field exists on given table and no row exists with given field with given value.
func isUnique(db *gorm.DB, table interface{}, varname string, val string) bool {
	var exists bool
	if err := db.Model(table).Select("count(*) > 0").Where(varname+" = ?", val).Find(&exists).Error; err != nil {
		return false
	} else {
		return !exists
	}
}

// Get the database tag for a struct.
func getJsonTag(v interface{}, structVar string) string {
	field, ok := reflect.TypeOf(v).Elem().FieldByName(structVar)
	if !ok {
		return ""
	} else {
		return field.Tag.Get("json")
	}
}

// Get database parameters string to place into DSN from a map.
func getDbParams(paramMap map[string]string) string {
	params := ""
	i := 0
	for key, val := range paramMap {
		if i > 0 {
			params += "&"
		}
		params += key + "=" + val
		i++
	}
	return params
}

// -- Validation

func isemail(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("Email must be a string.")
	}
	matcher := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !matcher.MatchString(st.String()) {
		return errors.New("Wrong email format")
	}
	parts := strings.Split(st.String(), "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return errors.New("Email server invalid!")
	}
	return nil
}

// Checks if a password contains upper case, lower case, numbers, and special characters.
func ispw(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("Value must be string!")
	} else {
		// Set password and character number.
		pw := st.String()
		restrictions := []string{"[a-z]", // Must contain lowercase.
			"^[" + A_NUMS + SPECIAL_CHARS + "]*$", // Must contain only some characters.
			"[A-Z]",                               // Must contain uppercase.
			"[0-9]",                               // Must contain numerics.
			"[" + SPECIAL_CHARS + "]"}             // Must contain special characters.
		for _, restriction := range restrictions {
			matcher := regexp.MustCompile(restriction)
			if !matcher.MatchString(pw) {
				return errors.New("Restriction not matched!")
			}
		}
		return nil
	}
}

// -- Password control --

// Hash a password
func hashPw(pw string) []byte {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pw), HASH_COST)
	return hash
}

// Compare password and hash for validity.
func comparePw(pw string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
