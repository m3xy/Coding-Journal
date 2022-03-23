package main

import (
	"log"
	"regexp"
	"strconv"
	"time"

	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	uuid "github.com/satori/go.uuid"
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
	USERTYPE_EDITOR             = 4

	// Password related
	HASH_COST = 8
)

var gormDb *gorm.DB
var DB_PARAMS map[string]string = map[string]string{
	"interpolateParams": "true",
	"parseTime":         "true",
}
var validate *validator.Validate

// User profile and personal information.
type User struct {
	ID           uint   `gorm:"primaryKey" json:"-"`
	GlobalUserID string `json:"-"`
	Email        string `gorm:"uniqueIndex;unique;not null" json:"email" validate:"email,required"`
	Password     string `gorm:"not null" json:"password,omitempty" validate:"min=8,max=64,ispw,required"`
	FirstName    string `json:"firstName" validate:"required,max=32"`
	LastName     string `json:"lastName" validate:"required,max=32"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	Organization string `json:"organization,omitempty"`

	CreatedAt time.Time      `json:",omitempty"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User global identification.
type GlobalUser struct {
	ID       string `gorm:"not null;primaryKey;type:varchar(191)" json:"userId" validate:"required"`
	UserType int    `gorm:"default:4" json:"userType"`
	User     *User  `json:"profile,omitempty"`

	AuthoredSubmissions []Submission `gorm:"many2many:authors_submission" json:"authoredSubmissions" validate:"dive"`
	ReviewedSubmissions []Submission `gorm:"many2many:reviewers_submission" json:"reviewedSubmissions" validate:"dive"`

	CreatedAt time.Time      `json:"createdAt"`
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
	// actual table fields
	gorm.Model
	Name     string `gorm:"not null;size:128;index" json:"name" validate:"max=118"`
	License  string `gorm:"size:64" json:"license" validate:"max=118"`
	Approved *bool  `json:"approved" gorm:"default:NULL"` // pointer to allow nil values as neither approved nor dissaproved

	// associations to other tables
	Files      []File       `json:"files,omitempty" validate:"dive"`
	Authors    []GlobalUser `gorm:"many2many:authors_submission" json:"authors,omitempty" validate:"required,dive"`
	Reviewers  []GlobalUser `gorm:"many2many:reviewers_submission" json:"reviewers,omitempty"`
	Categories []Category   `gorm:"many2many:categories_submissions" json:"categories,omitempty"` // tags for organizing/grouping code submissions (i.e. python)

	// stored in filesystem, not db
	MetaData *SubmissionData `gorm:"-" json:"metaData,omitempty"`
}

// structure for meta-data of the submission. matches the structure of the submission's
// JSON data file. This struct is never stored in the db
type SubmissionData struct {
	Abstract string    `json:"abstract"`
	Reviews  []*Review `json:"reviews"`
}

// struct for code files
type File struct {
	// stored in files table
	gorm.Model
	SubmissionID uint   `json:"submissionId"` // foreign key linking files and submissions tables
	Path         string `json:"path"`         // this path is relative from submission root
	// Name string `json:"name"`

	// association to other tables
	Comments []Comment `json:"comments,omitempty"`

	// stored in filesystem
	Base64Value string `gorm:"-" json:"base64Value"` // file content, only stored in filesystem
}

// Structure for submission reviews (not written to db, all in filesystem)
type Review struct {
	ReviewerID  string `json:"reviewerId"`
	Approved    bool   `json:"approved"`
	Base64Value string `json:"base64Value"`
}

// Structure for user comments on code
type Comment struct {
	gorm.Model
	AuthorID    string `json:"author"`
	FileID      uint   `json:"fileId"` // foreign key linking comments to files table
	Base64Value string `gorm:"type:mediumtext" json:"base64Value"`
	LineNumber  int    `json:"lineNumber"`

	// self association for replies to user comments
	ParentID *uint     `gorm:"default:NULL" json:"parentId,omitempty"` // pointer so it can be nil
	Comments []Comment `gorm:"foreignKey:ParentID" json:"comments,omitempty"`
}

// stores submission tags (i.e. networking, java, python, etc.)
// uniqueIndex:idx_first_second specifies the first and second column as a unique pair
type Category struct {
	Tag string `gorm:"primaryKey" json:"category"` // actual content of the tag

	CreatedAt time.Time `json:"-"`
	DeletedAt time.Time `json:"-"`
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
	err = db.AutoMigrate(&GlobalUser{}, &User{}, &Server{}, &Category{}, &Submission{}, &File{}, &Comment{})
	if err != nil {
		goto ERR
	}

	// Set up validation
	validate = validator.New()
	validate.RegisterValidation("ispw", ispw)
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

func getTagArray(categories []Category) []string {
	arr := []string{}
	for _, category := range categories {
		arr = append(arr, category.Tag)
	}
	return arr
}

// -- Validation -- //

// Checks if a password contains upper case, lower case, numbers, and special characters.
func ispw(fl validator.FieldLevel) bool {
	if fl.Field().String() == "invalid" {
		return false
	} else {
		// Set password and character number.
		pw := fl.Field().String()
		restrictions := []string{"[a-z]", // Must contain lowercase.
			"^[" + A_NUMS + SPECIAL_CHARS + "]*$", // Must contain only some characters.
			"[A-Z]",                               // Must contain uppercase.
			"[0-9]",                               // Must contain numerics.
			"[" + SPECIAL_CHARS + "]"}             // Must contain special characters.
		for _, restriction := range restrictions {
			matcher := regexp.MustCompile(restriction)
			if !matcher.MatchString(pw) {
				return false
			}
		}
		return true
	}
}
