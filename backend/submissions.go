// =============================================================================
// submissions.go
// Authors: 190010425
// Created: November 18, 2021
//
// TODO: write functionality for popularity statistics and an ordering algorithm
// for suggested submissions
//
// This file handles the reading/writing of all submissions (just the submission
// with its data, not the files themselves)
//
// Submission ID (as stored in db submissions table)
// 	> <submission_name>/ (as stored in the submissions table)
// 		... (submission directory structure)
// 	> .data/
// 		> submission_data.json
// 		... (submission directory structure)
// notice that in the filesystem, the .data dir structure mirrors the
// submission, so that each file in the submission can have a .json file storing
// its data which is named in the same way as the source code
// =============================================================================

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

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ENDPOINT_UPLOAD_SUBMISSION = "/create"
	SUBROUTE_SUBMISSION        = "/submission"
	ENDPOINT_SUBMISSIONS       = "/submissions"
)

// Describe mux routing for submission-based endpoints.
func getSubmissionsSubRoutes(r *mux.Router) {
	submissions := r.PathPrefix(ENDPOINT_SUBMISSIONS).Subrouter()
	submission := r.PathPrefix(SUBROUTE_SUBMISSION).Subrouter()
	submissions.Use(jwtMiddleware)

	// Submission routes:
	// + /submission/{id} - Get given submission.
	// + /submissions/create - Create a submission.
	submission.HandleFunc("/{id}", RouteGetSubmission).Methods(http.MethodGet)
	submissions.HandleFunc(ENDPOINT_UPLOAD_SUBMISSION, uploadSubmission).Methods(http.MethodPost, http.MethodOptions)
}

// ------
// Router Functions
// ------

// Router function to get the names and id's of every submission of a given user
// if no user id is given in the query parameters, return all valid submissions
//
// Response Codes:
//	200 : if the action completed successfully
// 	401 : if the proper security token was not given in the request
//	500 : otherwise
// Response Body:
//	A JSON object of form: {...<submission id>:<submission name>...}
func getAllAuthoredSubmissions(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] GetAllAuthoredSubmissions request received from %v", r.RemoteAddr)
	// gets the userID from the URL
	var userID string
	params := r.URL.Query()
	userIDs := params["authorID"]
	if userIDs == nil {
		userID = "*"
	} else {
		userID = userIDs[0]
	}

	// set content type for return
	w.Header().Set("Content-Type", "application/json")
	// uses getAuthoredSubmissions to get all user submissions by setting authorID = *
	submissions, err := getAuthoredSubmissions(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// marshals and returns the map as JSON
	jsonString, err := json.Marshal(submissions)
	if err != nil {
		log.Printf("[ERROR] JSON formatting failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// writes json string
	log.Printf("[INFO] GetAllSubmission request from %v successful", r.RemoteAddr)
	w.Write(jsonString)
}

// Router function to upload new submissions to the db. The body of the
// sent request should be a valid submission Json objects as specified
// in backend/README.md
func uploadSubmission(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] uploadSubmission request received from %v", r.RemoteAddr)

	// parses the Json request body into a submission struct
	resp := UploadSubmissionResponse{}
	reqBody := UploadSubmissionBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		resp.Message = "Incorrect submission fields."
		resp.Error = true
		w.WriteHeader(http.StatusBadRequest)
	} else if r.Context().Value("userId") == nil {
		// User is not validated - error out.
		resp.Message = "The client is unauthorized from making this query."
		resp.Error = true
		w.WriteHeader(http.StatusUnauthorized)
	} else if submissionID, err := ControllerUploadSubmission(reqBody); err != nil {

		// Respond according to the error type returned.
		switch err.(type) {
		case validator.ValidationErrors:
			resp.Message = fmt.Sprintf("Bad fields inserted - %v", err.(validator.ValidationErrors).Error())
			w.WriteHeader(http.StatusBadRequest)
		case *WrongPermissionsError:
			resp.Message = fmt.Sprintf("User %s does not have valid permissions.", err.(*WrongPermissionsError).userID)
			w.WriteHeader(http.StatusUnauthorized)
		case *BadUserError:
			resp.Message = fmt.Sprintf("User %s does not exist in the system.", err.(*BadUserError).userID)
			w.WriteHeader(http.StatusUnauthorized)
		default:
			resp.Message = "Internal server error - Undisclosed."
			w.WriteHeader(http.StatusInternalServerError)
		}
		resp.Error = true
	} else {
		resp.Message = "Submission creation successful!"
		resp.SubmissionID = submissionID
	}

	// Return response body after function successful.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !resp.Error {
		log.Print("[INFO] uploadSubmission request successful\n")
	}
}

// Decode a validated submission upload body into a submission model.
func ValidateAndDecodeSubmission(body UploadSubmissionBody) (*Submission, error) {
	// Get arrays of authors, reviewers and categories from a submission upload body.
	if err := validate.Struct(body); err != nil {
		return nil, err
	}

	// Get array of author and reviewer users from body.
	authors, reviewers := []GlobalUser{}, []GlobalUser{}
	for _, author := range body.Authors {
		authors = append(authors, GlobalUser{ID: author})
	}
	for _, reviewer := range body.Reviewers {
		reviewers = append(reviewers, GlobalUser{ID: reviewer})
	}

	return &Submission{
		Name: body.Name, License: body.License,
		Files: body.Files, Categories: body.Tags,
		Authors: authors, Reviewers: reviewers,
		MetaData: &SubmissionData{
			Abstract: "test abstract",
		},
	}, nil

}

// Controller for the upload submission POST request.
func ControllerUploadSubmission(body UploadSubmissionBody) (uint, error) {
	if submission, err := ValidateAndDecodeSubmission(body); err != nil {
		// Deal with validation errors - print all failed fields.
		if errors, ok := err.(validator.ValidationErrors); ok {
			log.Printf("[WARN] Given submission is invalid! ")
			return 0, errors
		} else {
			// Validation process failed.
			log.Printf("[ERROR] Validation failure! ")
			return 0, err
		}
	} else if submissionID, err := addSubmission(submission); err != nil {
		return 0, err
	} else {
		return submissionID, nil
	}
}

// Send submission data to the frontend for display. ID included for file
// and comment queries.
//
// Response Codes:
//	200 : if the submission exists and the request succeeded
// 	401 : if the proper security token was not given in the request
//	400 : if the request is invalid or badly formatted
// 	500 : if something else goes wrong in the backend
// Response Body: a submission object as specified in README.md
func RouteGetSubmission(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] getSubmission request received from %v", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// gets the submission ID from the URL parameters
	params := mux.Vars(r)
	var encodable interface{}

	// Check path, execute controller, check errors.
	submissionID64, err := strconv.ParseUint(params["id"], 10, 32)
	if err != nil {
		encodable = StandardResponse{Message: "Given ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)
	} else if submission, err := getSubmission(uint(submissionID64)); err != nil {
		switch err.(type) {
		case *NoSubmissionError: // The given submission doesn't exist
			encodable = StandardResponse{Message: err.(*NoSubmissionError).Error(), Error: true}
			w.WriteHeader(http.StatusNotFound)
		default: // Unexpected error - error out as server error.
			log.Printf("[ERROR] could not retrieve submission data properly: %v", err)
			encodable = StandardResponse{Message: "Internal server error - could not retrieve submission.", Error: true}
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	} else {
		encodable = submission
	}

	// writes JSON data for the submission to the HTTP connection
	if err := json.NewEncoder(w).Encode(encodable); err != nil {
		log.Printf("[ERROR] error formatting response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Print("[INFO] success\n")
	return
}

// ------
// Helper Functions
// ------

// Add submission to filesystem and database. All fields should be set.
// Authors and reviewers arrays only use GlobalUser.ID in this function
//
// Params:
//	submission (*Submission) : the submission to be added to the db
// 		(all fields but ID MUST be set)
// Returns:
//	(int) : the id of the added submission
//	(error) : if the operation fails
func addSubmission(submission *Submission) (uint, error) {
	// adds the submission to the db, automatically setting submission.ID
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Create(submission).Error; err != nil {
			return err
		} else if err := addAuthors(tx, submission.Authors, submission.ID); err != nil {
			return err
		} else if err := addReviewers(tx, submission.Reviewers, submission.ID); err != nil {
			return err
		} else if err := addTags(tx, submission.Categories, submission.ID); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return 0, err
	}

	// creates the directories to hold the submission in the filesystem
	submissionPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submission.ID), submission.Name)
	submissionDataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submission.ID), DATA_DIR_NAME, submission.Name)
	if err := os.MkdirAll(submissionPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	} else if err := os.MkdirAll(submissionDataPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	}

	// opens a JSON file for the submission metadata, and writes a SubmissionData struct to it
	dataFile, err := os.OpenFile(submissionDataPath+".json", os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
	if err != nil {
		return 0, err
	}
	defer dataFile.Close()
	if err := json.NewEncoder(dataFile).Encode(submission.MetaData); err != nil {
		return 0, err
	}

	// Adds each member file to the filesystem and database
	for _, file := range submission.Files {
		_, err = addFileTo(&file, submission.ID)
	}
	return submission.ID, nil
}

// Add an array of authors to the given submission provided the id given corresponds to a valid
// user with publisher or publisher-reviewer permissions
//
// Params:
//	authors ([]GlobalUser) : the global user structs of the authors to add to the submission. Only
// 		the GlobalUser.ID field must be set here
//	submissionID (int) : the id of the submission to be added to
// Returns:
//	(error) : an error if one occurs, nil otherwise
func addAuthors(tx *gorm.DB, authors []GlobalUser, submissionID uint) error {
	// Check if there is at least 1 author.
	switch {
	case authors == nil, len(authors) == 0:
		return errors.New("There must be at least 1 author")
	}
	return appendUsers(tx, authors, []int{USERTYPE_PUBLISHER, USERTYPE_REVIEWER_PUBLISHER}, "Authors", submissionID)
}

// Add an array of reviewers to the given submission provided the id given corresponds to a valid
// user with reviewer or publisher-reviewer permissions
//
// Params:
//	reviewers ([]GlobalUser) : the global user structs of the reviewers to add to the submission.
// 		Only GlobalUser.ID must be set here
//	submissionID (uint) : the id of the submission to be added to
// Returns:
//	(error) : an error if one occurs, nil otherwise
func addReviewers(tx *gorm.DB, reviewers []GlobalUser, submissionID uint) error {
	return appendUsers(tx, reviewers, []int{USERTYPE_REVIEWER, USERTYPE_REVIEWER_PUBLISHER}, "Reviewers", submissionID)
}

// Append users to submission at given association, with given priviledge restrictions.
func appendUsers(tx *gorm.DB, users []GlobalUser, priviledges []int, association string, submissionID uint) error {
	// No required appends. Skip.
	if users == nil {
		return nil
	}
	// Return transaction for user append with priviledge check
	return tx.Transaction(func(tx *gorm.DB) error {
		// For each user - find by ID and priviledges.
		for _, user := range users {
			if res := tx.Where("user_type IN ?", priviledges).Limit(1).Find(&user, "ID = ?", user.ID); res.Error != nil {
				return res.Error
			} else if res.RowsAffected == 0 {
				if isUnique(tx, &GlobalUser{}, "ID", user.ID) {
					return &BadUserError{userID: user.ID} // ID unique - user not registered.
				} else {
					return &WrongPermissionsError{userID: user.ID} // User registered - permissions false.
				}
			}
		}
		// Append checked users into the submission's association.
		submission := Submission{}
		submission.ID = submissionID
		if err := tx.Model(&submission).Association(association).Append(users); err != nil {
			return err
		}
		return nil
	})
}

// function to add tags for a given submission
//
// Parameters:
// 	tags ([]string) : the tags to add to the submission
// 	submissionID (int) : the unique ID of the submission to add tags to
// Returns:
// 	(error) : an error if one occurs, nil otherwise
func addTags(tx *gorm.DB, tags []string, submissionID uint) error {
	// Skip if no tags given.
	switch {
	case tags == nil:
		return nil
	case len(tags) == 0:
		return nil
	}

	// builds a list of category structs to be inserted
	categories := []Category{}
	for _, tag := range tags {
		if tag == "" {
			return errors.New("Tag cannot be an empty string")
		}
		categories = append(categories, Category{Tag: tag, SubmissionID: submissionID})
	}
	// inserts the tags using a transaction
	if err := tx.Transaction(func(tx *gorm.DB) error {
		// checks that the submission exists
		submission := &Submission{}
		submission.ID = submissionID
		if err := tx.Model(submission).First(submission).Error; err != nil {
			return err
		}
		// inserts the array of categories
		if err := tx.Model(&Category{}).Create(&categories).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// gets all submissions which are written by a given user and returns them
//
// Params:
// 	authorID (string) : the global id of the author as stored in the db
// Return:
// 	(map[int]string) : map of submission IDs to submission names
// 	(error) : an error if something goes wrong, nil otherwise
func getAuthoredSubmissions(authorID string) (map[uint]string, error) {
	// gets the author's submissions
	var submissions []*Submission
	if err := gormDb.Model(&GlobalUser{ID: authorID}).Association("AuthoredSubmissions").Find(&submissions); err != nil {
		return nil, err
	}
	// formats the authors submissions into a map of form submission.ID -> submission.Name
	subMap := make(map[uint]string)
	for _, sub := range submissions {
		subMap[sub.ID] = sub.Name
	}
	return subMap, nil
}

// gets all of the submissions which a given user is reviewing. Permissions
// are not checked here, as permissions get checked upon addition of a reviewer
//
// Parameters:
// 	reviewerID (string) : the global ID of the reviewer
// Returns:
// 	(map[int]string) : a map of form { <submission ID>:<submission name> }
// 	(error) : an error if something goes wrong, nil otherwise
func getReviewedSubmissions(reviewerID string) (map[uint]string, error) {
	// queries the submissions <-> reviewers association where the GlobalUser.ID = reviewerID
	var submissions []*Submission
	if err := gormDb.Model(&GlobalUser{ID: reviewerID}).Association("ReviewedSubmissions").Find(&submissions); err != nil {
		return nil, err
	}
	// formats the authors submissions into a map of form submission.ID -> submission.Name
	subMap := make(map[uint]string)
	for _, sub := range submissions {
		subMap[sub.ID] = sub.Name
	}
	return subMap, nil
}

// Get the submission struct corresponding to the id by querying the db and reading
// the submissions meta-data file. Note that an array of file objects is gotten here,
// but their content and metadata is not attached (this is queried via another function)
//
// Parameters:
// 	submissionID (int) : the submission's unique id
// Returns:
// 	(*Submission) : the data of the submission
// 	(error) : an error if one occurs
func getSubmission(submissionID uint) (*Submission, error) {
	// Get data contained inside the database.
	submission := &Submission{}
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		var err error
		if res := tx.Preload("Authors").Preload("Files").Preload("Reviewers").Find(submission, submissionID); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return &NoSubmissionError{ID: submissionID}
		}
		if submission.Categories, err = getSubmissionCategories(tx, submissionID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return &Submission{}, err
	}

	// gets the data which is not stored in the submissions table of the database
	var err error
	if submission.MetaData, err = getSubmissionMetaData(submissionID); err != nil {
		return nil, err
	}
	return submission, nil
}

// Queries the database for categories (tags) associated with the given
// submission (i.e. python, c++, sorting algorithm, etc.)
//
// Parameters:
// 	submissionID (int) : the submission to get the categories for
// Returns:
// 	([]string) : an array of the tags associated with the given submission
// 	(error) : an error if one occurs while retrieving the tags
func getSubmissionCategories(tx *gorm.DB, submissionID uint) ([]string, error) {
	// gets all tags with foreign key submission_id = submissionID
	var categories []Category
	if err := tx.Model(&Category{SubmissionID: submissionID}).Find(&categories).Error; err != nil {
		return nil, err
	}
	// loops over the query results to build a string array of tags
	tags := []string{}
	for _, category := range categories {
		tags = append(tags, category.Tag)
	}
	return tags, nil
}

// Queries the database for a submissions member files. Note that this
// function does not access the filesystem, and therefore does not return
// file content or metadata
//
// Params:
//	submissionID (int) : the id of the submission to get the files of
//Returns:
//	([]File) : Array of files which are members of the given submission
//	(error) : if something goes wrong during the query
func getSubmissionFiles(submissionID uint) ([]File, error) {
	var files []File
	file := &File{}
	file.SubmissionID = submissionID
	if err := gormDb.Model(file).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// This function gets a submission's meta-data from the filesystem
//
// Parameters:
// 	submissionID (int) : the unique id of the submission
// Returns:
//	(*SubmissionData) : the submission's metadata if found
// 	(error) : if anything goes wrong while retrieving the metadata
func getSubmissionMetaData(submissionID uint) (*SubmissionData, error) {
	// gets the submission name from the database
	submission := &Submission{}
	submission.ID = submissionID
	if err := gormDb.Model(submission).Select("submissions.Name").First(&submission).Error; err != nil {
		return nil, err
	}

	// reads the data file into a string
	dataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionID), DATA_DIR_NAME, submission.Name+".json")
	dataString, err := ioutil.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	// marshalls the string of data into a struct to be returned
	submissionData := &SubmissionData{}
	if err := json.Unmarshal(dataString, submissionData); err != nil {
		return nil, err
	}
	return submissionData, nil
}

// This function queries a submission in the local format from the db, and transforms
// it into the supergroup compliant format
//
// Parameters:
// 	submissionID (int) : the id of the submission to be converted to the supergroup-compliant
// 		format
// Returns:
// 	(*SupergroupSubmission) : a supergroup compliant submission struct
// 	(error) : an error if one occurs
func localToGlobal(submissionID uint) (*SupergroupSubmission, error) {
	// gets the submission struct which submissionID refers to
	localSubmission, err := getSubmission(submissionID)
	if err != nil {
		return nil, err
	}
	// creates the Supergroup metadata struct
	supergroupData := &SupergroupSubmissionData{
		CreationDate: fmt.Sprint(localSubmission.CreatedAt),
		Categories:   localSubmission.Categories,
		Abstract:     localSubmission.MetaData.Abstract,
		License:      localSubmission.License,
	}

	// adds author names to an array
	authorNames := []string{}
	for _, author := range localSubmission.Authors {
		authorNames = append(authorNames, author.FullName)
	}
	supergroupData.AuthorNames = authorNames

	// creates the list of file structs using the file paths and files.go
	var base64 string
	var supergroupFile *SupergroupFile
	supergroupFiles := []*SupergroupFile{}
	for _, file := range localSubmission.Files {
		fullFilePath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(localSubmission.ID), localSubmission.Name, file.Path)
		base64, err = getFileContent(fullFilePath)
		if err != nil {
			return nil, err
		}
		supergroupFile = &SupergroupFile{
			Name:        filepath.Base(file.Path),
			Base64Value: base64,
		}
		supergroupFiles = append(supergroupFiles, supergroupFile)
	}

	// creates the supergroup submission to return
	return &SupergroupSubmission{
		Name:     localSubmission.Name,
		Files:    supergroupFiles,
		MetaData: supergroupData,
	}, nil
}
