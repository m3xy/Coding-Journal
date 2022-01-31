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

type Submission struct {
	gorm.Model
	Name       string     `gorm:"not null;size:128;index" json:"submissionName"`
	Files      []File     `gorm:"foreignKey:SubmissionID;references:ID"`
	Authors    []User     `gorm:"many2many:authors_submission"`
	Reviewers  []User     `gorm:"many2many:reviewers_submission"`
	Categories []Category `gorm:"many2many:categories_submission"`
}

type File struct {
	gorm.Model
	SubmissionID uint
	FilePath     string `gorm:"unique" json:"filePath"`
}

type Category struct {
	gorm.Model
	Tag string `gorm:"uniqueIndex;unique" json:"category"`
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
	err = db.AutoMigrate(&User{}, &Server{}, &GlobalID{}, &Category{}, &Submission{}, &File{})
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
	tables := []interface{}{&File{}, &Category{}, &Submission{}, &GlobalID{}, &User{}}
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
