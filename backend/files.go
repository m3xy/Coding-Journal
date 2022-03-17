// ================================================================================
// files.go
// Authors: 190010425, 190014935
// Created: November 23, 2021
//
// This file handles reading/writing code files along with their
// data.
//
// The directory structure for the filesystem is as follows:
//
// Submission ID (as stored in db submissions table)
// 	> <submission_name>/ (as stored in the submissions table)
// 		... (submission directory structure)
// 	> .data/
// 		> submission_data.json
// 		... (submission directory structure)
// notice that in the filesystem, the .data dir structure mirrors the
// submission, so that each file in the submission can have a .json file storing
// its data which is named in the same way as the source code (the only difference
// being the extension)
// ================================================================================

package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	// Subroutes and endpoints for files
	SUBROUTE_FILE       = "/file"
	ENDPOINT_NEWFILE    = "/upload"
	ENDPOINT_NEWCOMMENT = "/comment"

	// File Mode Constants
	DIR_PERMISSIONS  = 0755 // permissions for filesystem directories
	FILE_PERMISSIONS = 0644 // permissions for submission files
)

func getFilesSubRoutes(r *mux.Router) {
	files := r.PathPrefix(SUBROUTE_FILE).Subrouter()
	files.Use(jwtMiddleware)

	// File subroutes:
	// + GET /file/{id} - Get given file.
	// + POST /file/{id}/comment - Post a new comment
	files.HandleFunc("/{id}", getFile).Methods(http.MethodGet)
	files.HandleFunc("/{id}"+ENDPOINT_NEWCOMMENT, uploadUserComment).Methods(http.MethodPost, http.MethodOptions)
}

// -----
// Router functions
// -----

// Get the path to the submissions directory.
func getSubmissionDirectoryPath(s Submission) string {
	return filepath.Join(FILESYSTEM_ROOT, fmt.Sprintf("%d-%d", s.ID, s.CreatedAt.Unix()))
}

// Retrieve code files from filesystem. Returns
// file with comments and metadata. Recieves a request
// with a file ID as a URL parameter
//
// Response Codes:
// 	200 : File exists, retrieved successfully
// 	401 : if the request does not have the proper security token
// 	400 : malformatted request, or non-existent file ID
// 	500 : if something else goes wrong in the backend
// Response Body:
// 	file: object
// 		ID: uint
//		CreatedAt: string
// 		UpdatedAt: string
// 		DeletedAt: string
// 		FilePath: string
// 		FileName: string
// 		SubmissionID: int
// 		Base64Value: string
// 		Comments: array
// 			Comment: object
// 				author: int
// 				CreatedAt: datetime string
// 				base64Value: string
// 				comments: array of objects (same as comments)
func getFile(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] getFile request received from %v", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// gets the fileID from the URL parameters as uint. Must unwrap from uint64
	params := mux.Vars(r)
	fileID64, err := strconv.ParseUint(params["id"], 10, 32) // specifies width as 32
	if err != nil {
		log.Printf("[ERROR] FileID: %s unable to be parsed", params["id"])
		w.WriteHeader(http.StatusBadRequest) // TODO: maybe use GOTO here
		return
	}
	fileID := uint(fileID64) // gets uint from uint64

	// gets the file data from the db and filesystem
	file, err := getFileData(fileID)
	if err != nil {
		log.Printf("[ERROR] unable to get file data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// writes JSON data for the file to the HTTP connection if no error has occured
	response, err := json.Marshal(file)
	if err != nil {
		log.Printf("[ERROR] JSON formatting failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] getFile request from %v successful.", r.RemoteAddr)
	w.Write(response)
}

// -----
// Helper Functions
// -----

// Add file into submission, and store it to filesystem and database
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
		if err := gormDb.Model(file).Find(file).Error; err != nil {
			return err
		}
		// gets the file comments
		// Order comments by newest.
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

// Recursive functions for loading replies.
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

// Get file content from filesystem.
// Params:
// 	filePath (string): an absolute path to the file
// Returns:
// 	(error) : if something goes wrong, nil otherwise
func getFileContent(filePath string) (string, error) {
	// reads in the file's content
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// if no error occurred, assigns file.Base64Value a value
	return string(fileData), nil
}

// Unzip a zip file into a file array, from the zip's base 64.
// Returns the array of files.
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
			Base64Value: base64.URLEncoding.EncodeToString(buf.Bytes()),
		}
	}
	return files, nil
}

// Unzip a file to some temporary folder. Returns folder path.
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
