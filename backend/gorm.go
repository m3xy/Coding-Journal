package main

import (
	"log"
	"time"

	"fmt"
	"os"

	"github.com/satori/go.uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;->;default:uuid()" json:"userId"`
	Email        string    `gorm:"uniqueIndex;unique;not null" json:"email"`
	Password     string    `gorm:"not null" json:"password" validate:"min=8,max=64,validpw"`
	FirstName    string    `gorm:"index:idx_user,class:FULLTEXT" validate:"nonzero,max=32" json:"firstname"`
	LastName     string    `gorm:"index:idx_user,class:FULLTEXT" validate:"nonzero,max=32" json:"lastname"`
	UserType     int       `json:"usertype"`
	PhoneNumber  string    `gorm:"index:,class:FULLTEXT" json:"phonenumber"`
	Organization string    `gorm:"index:,class:FULLTEXT" json:"organization"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Server struct {
	gorm.Model
	GroupNumber int    `gorm:"not null;primaryKey"`
	Token       string `gorm:"size:1028;not null"`
	Url         string `gorm:"not null; size:512"`
}

type GlobalID struct {
	gorm.Model
	GlobalID string `gorm:"not null;primaryKey", json:"globalId"`
	User     User
}

type File struct {
	gorm.Model
	FilePath string `gorm:"unique", json:"filePath"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `gorm:"index"`
}

type Submission struct {
	gorm.Model
	Name       string `gorm:"not null;size:128;index" json:"submissionName"`
	Files      []File
	Authors    []User
	Reviewers  []User
	Categories []Category

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `gorm:"index"`
}

type Category struct {
	gorm.Model
	Tag string `gorm:"uniqueIndex;unique" json:"category"`
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	user.ID = uuid.NewV4()
	return
}

func gormInit(dbname string) (*gorm.DB, error) {
	// Set MySQL info in DSN format according to Go MySQL Drive -
	// user:password@protocol(host:port)/dbname?[param1=val...]
	mysqlInfo := fmt.Sprintf("%s/%s?%s", os.Getenv("DATABASE_URL"), dbname,
		getDbParams(DB_PARAMS)) // Setting this to allow prepared statements.
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{})
	if err != nil {
		goto ERR
	}
	err = db.AutoMigrate(&User{}, &Server{}, &GlobalID{}, &File{}, &Submission{}, &Category{})
	if err != nil {
		goto ERR
	}
	return db, nil

ERR:
	log.Fatalf("SQL initialization error: %v", err)
	return nil, err
}
