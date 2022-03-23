// =================================================================
// files.go
// Authors: 190010425, 190014935
// Created: November 23, 2021
//
// This file handles reading/writing code files along with their
// data.
// =================================================================

package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// file constants, includes
const (
	// TEMP: hard coded for testing. Adapt to using an environment variable
	FILESYSTEM_ROOT = "../filesystem/" // path to the root directory holding all submission directories
	DATA_DIR_NAME   = ".data"          // name of the hidden data dir to be put into the submission directory structure

	SUBROUTE_FILE       = "/file"
	ENDPOINT_NEWFILE    = "/upload"
	ENDPOINT_NEWCOMMENT = "/comment"

	DIR_PERMISSIONS  = 0755 // permissions for filesystem directories
	FILE_PERMISSIONS = 0644 // permissions for submission files
)

func getFilesSubRoutes(r *mux.Router) {
	files := r.PathPrefix(SUBROUTE_FILE).Subrouter()
	files.Use(jwtMiddleware)

	// File subroutes:
	// + GET /file/{id} - Get given file.
	// + POST /file/{id}/comment - Post a new comment
	files.HandleFunc("/{id}", GetFile).Methods(http.MethodGet)
	files.HandleFunc("/{id}"+ENDPOINT_NEWCOMMENT, PostUploadUserComment).Methods(http.MethodPost, http.MethodOptions)
}

// Get the path to the submissions directory.
func getSubmissionDirectoryPath(s Submission) string {
	return filepath.Join(FILESYSTEM_ROOT, fmt.Sprintf("%d-%d", s.ID, s.CreatedAt.Unix()))
}

// -----
// Router functions
// -----

// Returns file with comments and metadata.
// GET /file/{id}
func GetFile(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] getFile request received from %v", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	var err error
	resp := &GetFileResponse{}

	// gets the fileID from the URL parameters as uint. Must unwrap from uint64
	params := mux.Vars(r)
	fileID64, err := strconv.ParseUint(params["id"], 10, 32) // specifies width as 32
	if err != nil {
		resp.StandardResponse = StandardResponse{Message: "Bad Request - file ID unable to be parsed", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// calls helper function to get the file struct for the given ID
	} else if resp.File, err = getFileData(uint(fileID64)); err != nil {
		switch err.(type) {
		case *FileNotFoundError:
			resp.StandardResponse = StandardResponse{Message: "Bad Request - no file exists for the given ID", Error: true}
			w.WriteHeader(http.StatusBadRequest)
		default:
			resp.StandardResponse = StandardResponse{Message: "Internal Server Error - undisclosed", Error: true}
			log.Printf("[ERROR] unable to get file data: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// Encode response - set as error if empty
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] JSON repsonse formatting failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Printf("[INFO] getFile request from %v successful.", r.RemoteAddr)
	}
}

// -----
// Helper Functions
// -----

// Add file to submission, and store it in filesystem and database
// Note: Need valid submission. No comments exist on file
// creation.
//
// Params:
// 	file (*File) : the file to add to the db and filesystem (all fields but ID and SubmissionID MUST be set)
// 	submissionID (int) : the id of the submission which the added file is to be linked
// 		to as an unsigned integer
// Returns:
// 	(int) : the id of the added file (0 if an error occurs)
// 	(error) : if the operation fails
func addFileTo(file *File, submissionID uint) (uint, error) {
	// inserts the file into the db, and gets the submission name
	submission := &Submission{}
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// queries the submission name (for use in accessing the filesystem)
		if err := tx.Select("Name, created_at, ID").First(submission, submissionID).Error; err != nil {
			return err
		}
		// adds a file to the submission in the db provided the submission exists
		if err := tx.Model(submission).Association("Files").Append(file); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}

	// Add file to filesystem creating dirs if you do not exist
	filePath := filepath.Join(getSubmissionDirectoryPath(*submission), fmt.Sprint(file.ID))

	// Create file path from its ID
	if codeFile, err := os.Create(filePath); err != nil {
		return 0, err
	} else {
		defer codeFile.Close()
		codeFile.Write([]byte(file.Base64Value))
	}

	return file.ID, nil
}

// helper function to return a file object given its ID
//
// Params:
// 	fileID (int) : the file's unique id
// Returns:
//	(*File) : the a file struct corresponding to the given ID
// 	(error) : an error if something goes wrong
func getFileData(fileID uint) (*File, error) {
	submission := &Submission{}
	file := &File{}
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// queries the file from the database
		file.ID = fileID
		if res := gormDb.Model(file).Find(file); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return &FileNotFoundError{ID: fileID}
		}
		// gets the file comments ordered by newest.
		var comments []Comment
		if err := gormDb.Model(&Comment{}).Order("created_at desc").
			Preload("Comments").Where("comments.parent_id IS NULL").
			Find(&comments, "file_id = ?", fileID).Error; err != nil {
			return err
		}
		for _, comment := range comments {
			if err := loadComments(tx, comment); err != nil {
				return err
			}
		}
		file.Comments = comments

		// queries the submission name
		if err := gormDb.Select("Name, ID, created_at").Find(submission, file.SubmissionID).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// builds path to the file using the queried submission name
	fullFilePath := filepath.Join(getSubmissionDirectoryPath(*submission), fmt.Sprint(file.ID))

	// gets file content
	var err error
	file.Base64Value, err = getFileContent(fullFilePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Recursive functions for loading comment replies.
// 
// Params:
// 	tx (*gorm.DB) - A gorm.DB instance to be query on (as this is used in transactions only)
// 	c (Comment) - the comment to get the children of
// Returns:
// 	(error) - an error if one occurs
func loadComments(tx *gorm.DB, c Comment) error {
	for _, child := range c.Comments {
		if err := tx.Preload("Comments").Order("created_at desc").
			Where("comments.parent_id = ?", c.ID).Find(&child).Error; err != nil {
			return err
		} else if err := loadComments(tx, child); err != nil {
			return err
		}
	}
	return nil
}

// Get base64 encoded file content from filesystem.
//
// Params:
// 	filePath (string): an absolute path to the file
// Returns:
// 	(error) : if something goes wrong, nil otherwise
func getFileContent(filePath string) (string, error) {
	// reads in the file's content
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// if no error occurred, assigns file.Base64Value a value
	return string(fileData), nil
}

// Unzip a zip file into a file array, from the zip's base 64.
//
// Params:
// 	base64value (string) : the base64 value of the entire zip file
// Returns:
// 	([]File) : the file array contained within the zip
// 	(error) : and error if one occurs
func getFileArrayFromZipBase64(base64value string) ([]File, error) {
	// Decode zip to temporary file for reading.
	var reader *zip.ReadCloser
	zipPath, err := TmpStoreZip(base64value)
	if err != nil {
		return nil, err
	} else if reader, err = zip.OpenReader(zipPath); err != nil {
		log.Printf("[ERROR] Failed to open reader for zip file - %v", err)
		os.Remove(zipPath)
		return nil, err
	}
	defer os.Remove(zipPath)
	defer reader.Close()

	// Iterate file-per-file unzip.
	files := make([]File, len(reader.File))
	for i, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			log.Printf("[ERROR] Failed to open file contained in the zip - %v", err)
			return nil, err
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(rc)
		files[i] = File{
			Path:        file.FileHeader.Name,
			Base64Value: base64.StdEncoding.EncodeToString(buf.Bytes()),
		}
	}
	return files, nil
}

// Unzip a file to some temporary folder. Returns folder path.
// 
// Params:
// 	base64value (string) : the base64 encoded value of the entire zip file
// Returns:
// 	(string) : the path to the file
// 	(error) : an error if one occurs
func TmpStoreZip(base64value string) (string, error) {
	zipBytes, err := base64.StdEncoding.DecodeString(base64value)

	if err != nil {
		log.Printf("[ERROR] Base 64 value given is invalid/corrupt.")
		return "", err
	}
	f, err := os.CreateTemp("/tmp", "*.zip")
	if err != nil {
		log.Printf("[ERROR] Cannot create temp file! %v", err)
		return "", err
	}
	defer f.Close()
	path := f.Name()
	if err := os.WriteFile(path, zipBytes, 0666); err != nil {
		log.Printf("[ERROR] ZIP file creation failed: %v", err)
		goto ROLLBACK
	}
	log.Printf("[INFO] Created ZIP file at path %s", path)
	return path, nil

ROLLBACK:
	os.Remove(path)
	return "", err
}
