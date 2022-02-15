// ================================================================================
// files.go
// Authors: 190010425
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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// file constants, includes
const (
	// TEMP: hard coded for testing. Adapt to using an environment variable
	FILESYSTEM_ROOT = "../filesystem/" // path to the root directory holding all submission directories
	DATA_DIR_NAME   = ".data"          // name of the hidden data dir to be put into the submission directory structure

	// File Mode Constants
	DIR_PERMISSIONS  = 0755 // permissions for filesystem directories
	FILE_PERMISSIONS = 0644 // permissions for submission files
)

func getFilesSubRoutes(r *mux.Router) {
	files := r.PathPrefix("/").Subrouter()
	files.HandleFunc(ENDPOINT_FILE, getFile).Methods(http.MethodGet, http.MethodOptions)
	files.HandleFunc(ENDPOINT_NEWCOMMENT, uploadUserComment).Methods(http.MethodPost, http.MethodOptions)
}

// -----
// Router functions
// -----

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
// 				replies: object (same as comments)
func getFile(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] getFile request received from %v", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// gets the fileID from the URL parameters as uint. Must unwrap from uint64
	params := r.URL.Query()
	fileID64, err := strconv.ParseUint(params["id"][0], 10, 32) // specifies width as 32
	if err != nil {
		log.Printf("[ERROR] FileID: %s unable to be parsed", params["id"][0])
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

// upload comment router function. Takes in a POST request and
// uses it to add a comment to the given file
//
// Response Codes:
// 	200 : comment was added succesfully
// 	401 : if the request does not have the proper security token
// 	400 : if the comment was not sent in the proper format
// 	500 : if something else goes wrong in the backend
// Response Body: empty
func uploadUserComment(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] uploadUserComment request received from %v.", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// gets the fileID and authorID from the URL parameters
	params := r.URL.Query()
	authorID := params["authorID"][0]
	fileID64, err := strconv.ParseUint(params["fileID"][0], 10, 32)
	if err != nil {
		log.Printf("[ERROR] FileID: %s unable to be parsed", params["id"][0])
		w.WriteHeader(http.StatusBadRequest) // TODO: maybe use GOTO here
		return
	}
	fileID := uint(fileID64) // gets uint from uint64

	// parses the json request body into a map (should just contain base64 comment content)
	var request map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("[ERROR] JSON decoding failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Insert data into Comment structure
	comment := &Comment{
		AuthorID:    authorID, // authors user id
		CreatedAt:        fmt.Sprint(time.Now()),
		Base64Value: request[getJsonTag(&Comment{}, "Base64Value")].(string),
		Replies:     nil, // replies are nil upon insertion
	}

	// adds the comment to the file, returns code OK if successful
	if err = addComment(comment, fileID); err != nil {
		log.Printf("[ERROR] Comment creation failed: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] uploadUserComment request from %v successful.", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
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
	// error cases 
	if file.Name == "" {
		return 0, errors.New("File name must be set")
	}

	// inserts the file into the db, and gets the submission name
	submission := &Submission{}
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// queries the submission name (for use in accessing the filesystem)
		submission.ID = submissionID
		if err := gormDb.Model(submission).Select("submissions.name").First(submission).Error; err != nil {
			return err
		}
		// adds a file to the submission in the db provided the submission exists
		if err := gormDb.Model(submission).Association("Files").Append(file); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}

	// Add file to filesystem
	filePath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionID), submission.Name, file.Path)
	fileDataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionID), DATA_DIR_NAME,
		submission.Name, strings.Replace(file.Path, filepath.Ext(file.Path), ".json", 1))

	// file paths without the file name (to create dirs if they don't exist yet)
	fileDirPath := filepath.Dir(filePath)
	fileDataDirPath := filepath.Dir(fileDataPath)

	// creates all directories on the file's relative path in case any of them do not exist yet
	if err := os.MkdirAll(fileDirPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	} else if err := os.MkdirAll(fileDataDirPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	}

	// Create and open content file and it's corresponding data file
	codeFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
	if err != nil {
		return 0, err
	}
	dataFile, err := os.OpenFile(
		fileDataPath, os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
	if err != nil {
		return 0, err
	}

	// write the file content
	if _, err = codeFile.Write([]byte(file.Base64Value)); err != nil {
		return 0, err
	}
	// Writes given file's metadata
	jsonString, err := json.Marshal(file.MetaData)
	if err != nil {
		return 0, err
	}
	if _, err = dataFile.Write([]byte(jsonString)); err != nil {
		return 0, err
	}

	// closes files
	codeFile.Close()
	dataFile.Close()

	return file.ID, nil
}

// Add a comment to a given file. Currently this can only append comments
//
// Params:
//	comment (*Comment) : The comment struct to add to the file
//	fileID (uint) : the id of the file to add a comment to
// Returns:
//	(error) : an error if one occurs, nil otherwise
func addComment(comment *Comment, fileID uint) error {
	// error cases
	if comment == nil {
		return errors.New("Comment cannot be nil")
	}

	// checks that the author of the comment is a registered user (either here or in another journal)
	author := &GlobalUser{}
	author.ID = comment.AuthorID
	if err := gormDb.Model(author).Omit("*").First(author).Error; err != nil {
		return err
	}

	// gets data to build the file path
	file := &File{}
	submission := &Submission{}
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// queries the file from the database
		file.ID = fileID
		if err := gormDb.Model(file).Select("files.name", "files.path", "files.submission_id").Find(file).Error; err != nil {
			return err
		}
		// queries the submission name
		submission.ID = file.SubmissionID
		if err := gormDb.Model(submission).Select("submissions.name").Find(submission).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	// builds the path to the metadata file
	fullDataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(file.SubmissionID), DATA_DIR_NAME,
		submission.Name, strings.Replace(file.Path, filepath.Ext(file.Path), ".json", 1))

	// gets the file's metadata, and appends the new comment to it
	fileData, err := getFileMetaData(fullDataPath)
	if err != nil {
		return err
	}
	// adds the new comment and writes to the metadata file
	fileData.Comments = append(fileData.Comments, comment)
	dataBytes, err := json.Marshal(fileData)
	if err != nil {
		return err
	}
	// if everything has gone correctly, the new data is written to the file
	return ioutil.WriteFile(fullDataPath, dataBytes, FILE_PERMISSIONS)
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
		// queries the submission name
		submission.ID = file.SubmissionID
		if err := gormDb.Model(submission).Select("submissions.name").Find(submission).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// builds path to the file and it's corresponding data file using the queried submission name
	fullFilePath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(file.SubmissionID), submission.Name, file.Path)
	fullDataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(file.SubmissionID), DATA_DIR_NAME,
		submission.Name, strings.Replace(file.Path, filepath.Ext(file.Path), ".json", 1))

	// gets file content and metadata
	var err error
	file.Base64Value, err = getFileContent(fullFilePath)
	if err != nil {
		return nil, err
	}
	file.MetaData, err = getFileMetaData(fullDataPath)
	if err != nil {
		return nil, err
	}
	return file, nil
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

// Get a file's metadata from filesystem, and returns a FileData
// objects
//
// Params:
//	dataPath (string) : a path to the data file containing a given file's meta-data
//Returns:
//  (*FileData) : the file's metadata
//	(error) : if something goes wrong, nil otherwise
func getFileMetaData(dataPath string) (*FileData, error) {
	// reads the file contents into a json string
	jsonData, err := ioutil.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}
	// fileData is parsed from json into the FileData struct
	fileData := &FileData{}
	err = json.Unmarshal(jsonData, fileData)
	if err != nil {
		return nil, err
	}
	// if no error occurred, return metadata
	return fileData, nil
}
