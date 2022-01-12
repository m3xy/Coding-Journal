// =============================================================================
// submissions.go
// Authors: 190010425
// Created: November 18, 2021
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
// its data which is named in the same way as the source code (the only difference
// being the extension)
// =============================================================================

package main

import (
	"database/sql"
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
)

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
func getAllSubmissions(w http.ResponseWriter, r *http.Request) {
	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
		return
	}
	if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("[INFO] GetAllSubmissions request received from %v", r.RemoteAddr)
	// gets the userId from the URL
	var userId string
	params := r.URL.Query()
	userIds := params[getJsonTag(&Credentials{}, "Id")]
	if userIds == nil {
		userId = "*"
	} else {
		userId = userIds[0]
	}

	// set content type for return
	w.Header().Set("Content-Type", "application/json")
	// uses getUserSubmissions to get all user submissions by setting authorId = *
	submissions, err := getUserSubmissions(userId)
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
	log.Printf("[INFO] getAllSubmission request from %v successful", r.RemoteAddr)
	w.Write(jsonString)
}

// Get submission for display on frontend. ID included for file and comment queries.
//
// Response Codes:
//	200 : if the submission exists and the request succeeded
// 	401 : if the proper security token was not given in the request
//	400 : if the request is invalid or badly formatted
// 	500 : if something else goes wrong in the backend
// Response Body:
// 	submission: Object
//		id: String
//		name: String
//		authors: Array of Author userIDs
// 		reviewers: Array of Reviewer userIDs
// 		files: Array of file path strings
//		metadata: object
//			abstract: String
// 			reviews: array of objects
// 				author: int
//				time: datetime string
//				content: string
//				replies: object (same as comments)
func getSubmission(w http.ResponseWriter, r *http.Request) {
	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
		return
	}
	if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("getSubmission request received from %v", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	// gets the submission Id from the URL parameters
	params := r.URL.Query()
	submissionId, err := strconv.Atoi(params[getJsonTag(&Submission{}, "Id")][0])
	if err != nil {
		log.Printf("invalid id: %s\n", params[getJsonTag(&Submission{}, "Id")][0])
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// creates a submission with the given ID
	submission := &Submission{Id: submissionId}

	// statement to query submission name
	getSubmissionName := fmt.Sprintf(SELECT_ROW,
		getDbTag(&Submission{}, "Name"),
		TABLE_SUBMISSIONS,
		getDbTag(&Submission{}, "Id"),
	)
	// executes query
	row := db.QueryRow(getSubmissionName, submissionId)
	if err := row.Scan(&submission.Name); err != nil {
		goto sqlerror
	}

	// gets submission authors, reviewers, and file paths (as relpaths from the root of the submission)
	submission.Authors, err = getSubmissionAuthors(submissionId)
	if err != nil {
		goto sqlerror
	}
	submission.Reviewers, err = getSubmissionReviewers(submissionId)
	if err != nil {
		goto sqlerror
	}
	submission.FilePaths, err = getSubmissionFiles(submissionId)
	if err != nil {
		goto sqlerror
	}
	submission.MetaData, err = getSubmissionMetaData(submissionId)
	if err != nil {
		goto sqlerror
	}
	goto success

sqlerror:
	log.Printf("[ERROR] getSubmissions SQL query error: %v", err)
	w.WriteHeader(http.StatusInternalServerError)
	return

success:
	// writes JSON data for the submission to the HTTP connection
	response, err := json.Marshal(submission)
	if err != nil {
		log.Printf("error formatting response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Print("success\n")
	w.Write(response)
	return
}

// Router function to import Journal submissions (submissions) from other journals
//
// TODO: fix this function
//
// Responses:
// 	- 200 : if the action completed successfully
// 	- 400 : if the request is badly formatted
// 	- 500 : if something goes wrong on our end
func importFromJournal(w http.ResponseWriter, r *http.Request) {
	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
		return
	}
	if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("importFromJournal request received from %v", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	// parses the data into a structure
	var importedSubmission *SupergroupSubmission
	err := json.NewDecoder(r.Body).Decode(importedSubmission)
	if err != nil {
		log.Printf("[ERROR] submission decoding failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// moves the data from the structure into a locally compliant format
	// don't receive user email, so it creates a fake one
	authorEmail := fmt.Sprintf("%s@email.com",
		strings.Replace(importedSubmission.Metadata.AuthorName, " ", "_", 4))
	author := &Credentials{
		Fname:    strings.Split(importedSubmission.Metadata.AuthorName, " ")[0],
		Lname:    strings.Split(importedSubmission.Metadata.AuthorName, " ")[1],
		Email:    authorEmail,
		Pw:       "password", // defaults to password here as we have no way of identifying users
		Usertype: USERTYPE_PUBLISHER,
	}
	authorId, err := registerUser(author)
	// formats the data in a submission
	submission := &Submission{
		Name:      strings.Replace(importedSubmission.Name, " ", "_", 10), // default is 10 spaces to replace
		Reviewers: []string{},
		Authors:   []string{authorId},
		FilePaths: []string{},
		MetaData: &CodeSubmissionData{ // TODO figure this out
			Abstract: "",
			Reviews:  []*Comment{},
		},
	}

	// adds the submission
	submissionId, err := addSubmission(submission)
	if err != nil {
		log.Printf("[ERROR] submission import failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	submission.Id = submissionId

	// adds the file
	for _, submissionFile := range importedSubmission.Files {
		file := &File{
			SubmissionId:   submissionId,
			SubmissionName: importedSubmission.Name,
			Name:           submissionFile.Name,
			Path:           submissionFile.Name,
			Content:        submissionFile.Content,
		}
		_, err = addFileTo(file, submissionId)
		if err != nil {
			log.Printf("[ERROR] file import to submission %d failed: %v", submissionId, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	// writes status OK if nothing goes wrong
	log.Printf("[INFO] importFromJournal request from %v successful.", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
}

// ------
// Helper Functions
// ------

// Add submission to filesystem and database.
// Note: submission ID is set by this function.
// Params:
//	submission (*Submission) : the submission to be added to the db (all fields but Id MUST be set)
// Returns:
//	(int) : the id of the added submission
//	(error) : if the operation fails
func addSubmission(submission *Submission) (int, error) {
	// error cases
	if submission == nil {
		return 0, errors.New("Submission cannot be nil")
	} else if submission.Name == "" {
		return 0, errors.New("Submission.Name must be set to a valid string")
	} else if submission.Authors == nil {
		return 0, errors.New("Authors array cannot be nil")
	} else if submission.Reviewers == nil {
		return 0, errors.New("Reviewers array cannot be nil")
	}

	// declares return values
	var submissionId int
	var err error

	// formats query to insert the submission into the db
	insertSubmission := fmt.Sprintf(
		INSERT_PROJ, TABLE_SUBMISSIONS,
		getDbTag(&Submission{}, "Name"))

	// executes the query and gets the submission id
	row := db.QueryRow(insertSubmission, submission.Name)
	if row.Err() != nil {
		return 0, row.Err()
	}
	// gets the id from the inserted submission
	if err = row.Scan(&submissionId); err != nil {
		return 0, err
	}

	// adds the authors and reviewers to their respective tables
	// (here we work with the assumption that author and reviewer arrays are very small)
	for _, authorId := range submission.Authors {
		if err = addAuthor(authorId, submissionId); err != nil {
			return 0, err
		}
	}
	for _, reviewerId := range submission.Reviewers {
		if err = addReviewer(reviewerId, submissionId); err != nil {
			return 0, err
		}
	}

	// adds a submission to the filesystem
	submissionPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), submission.Name)
	submissionDataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), DATA_DIR_NAME, submission.Name)
	if err = os.MkdirAll(submissionPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	}
	if err = os.MkdirAll(submissionDataPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	}

	// inserts the submission's metadata into the filesystem
	dataFile, err := os.OpenFile(
		submissionDataPath+".json", os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
	if err != nil {
		return 0, err
	}
	// Write submission metadata to the created file
	jsonString, err := json.Marshal(submission.MetaData)
	if err != nil {
		return 0, err
	}
	if _, err = dataFile.Write([]byte(jsonString)); err != nil {
		return 0, err
	}
	dataFile.Close()

	// if the action was successful, the submission id of the submission struct is set and returned
	submission.Id = submissionId
	return submissionId, nil
}

// Add an author to the given submission provided the id given corresponds to a valid
// user with publisher or publisher-reviewer permissions
//
// Params:
//	authorId (int) : the global id of the author to add to the submission
//	submissionId (int) : the id of the submission to be added to
// Returns:
//	(error) : an error if one occurs, nil otherwise
func addAuthor(authorId string, submissionId int) error {
	if submissionId < 0 {
		return errors.New(
			fmt.Sprintf("Submission IDs must be integers 0 or greater, not: %d", submissionId))
	}

	// checks that the author is a valid user with publisher or publisher-reviewer permissions
	var permissions int
	queryUserType := fmt.Sprintf(
		SELECT_ROW,
		getDbTag(&Credentials{}, "Usertype"),
		VIEW_PERMISSIONS,
		getDbTag(&IdMappings{}, "GlobalId"),
	)
	// executes the query, getting the user's permissions
	row := db.QueryRow(queryUserType, authorId)
	if err := row.Scan(&permissions); err != nil {
		return err
	}

	// checks permissions, and if they are correct, the author is added
	if permissions == USERTYPE_PUBLISHER || permissions == USERTYPE_REVIEWER_PUBLISHER {
		_, err := db.Query(
			fmt.Sprintf(INSERT_AUTHOR, TABLE_AUTHORS), submissionId, authorId)
		return err
	} else {
		return errors.New("User must be authorized as Publisher " +
			"to be listed as submission Author" + fmt.Sprint(permissions))
	}
}

// Add a user to a submission as a reviewer
//
// Params:
//	reviewerId (int) : the id of the reviewer to add to the submission
//	submissionId (int) : the id of the submission to be added to
// Returns:
//	(error) : an error if one occurs, nil otherwise
func addReviewer(reviewerId string, submissionId int) error {
	var err error
	if submissionId < 0 {
		return errors.New(fmt.Sprintf("Submission IDs must be integers 0 or greater, not: %d", submissionId))
	}

	// checks that the reviewer is a valid user with reviewer or publisher-reviewer permissions
	var permissions int
	queryUserType := fmt.Sprintf(
		SELECT_ROW,
		getDbTag(&Credentials{}, "Usertype"),
		VIEW_PERMISSIONS,
		getDbTag(&IdMappings{}, "GlobalId"),
	)
	// executes the query, only returning one row
	row := db.QueryRow(queryUserType, reviewerId)
	if row.Err() != nil {
		return row.Err()
	}
	// gets the user's permissions
	if err := row.Scan(&permissions); err != nil {
		return err
	}

	// checks permissions, and if they are correct, the reviewer is added
	if permissions != USERTYPE_REVIEWER && permissions != USERTYPE_REVIEWER_PUBLISHER {
		return errors.New("User must be authorized as a Reviewer")
	} else {
		_, err = db.Query(fmt.Sprintf(INSERT_REVIEWER, TABLE_REVIEWERS), submissionId, reviewerId)
		return err
	}
}

// gets all authors which are written by a given user and returns them
//
// Params:
// 	authorId (string) : the global id of the author as stored in the db
// Return:
// 	(map[int]string) : map of submission Ids to submission names
// 	(error) : an error if something goes wrong, nil otherwise
func getUserSubmissions(authorId string) (map[int]string, error) {
	// queries the database for the submission ID and name pairs
	columns := fmt.Sprintf("%s, %s",
		getDbTag(&AuthorsReviewers{}, "SubmissionId"),
		getDbTag(&Submission{}, "Name"),
	)
	stmt := fmt.Sprintf(SELECT_ROW, columns, VIEW_SUBMISSIONLIST, "userId")
	rows, err := db.Query(stmt, authorId)
	if err != nil {
		log.Printf("[WARN] User does exist: %s", authorId)
		return nil, err
	}

	// parses query result into { id : submission name } mappings
	var id int
	var submissionName string
	submissions := make(map[int]string)
	for rows.Next() {
		if err := rows.Scan(&id, &submissionName); err != nil {
			if err == sql.ErrNoRows {
				log.Printf("[WARN] Submission does not exist: %d", id)
				return nil, nil
			}
			log.Printf("[ERROR] SQL Error on submission retrieval.")
			return nil, err
		}
		submissions[id] = submissionName
	}
	return submissions, nil
}

// Query the authors of a given submission from the database
//
// Params:
//	submissionId (int) : the id of the submission to get authors of
// Returns:
//	[]string of the author's names
//	error if something goes wrong during the query
func getSubmissionAuthors(submissionId int) ([]string, error) {
	// builds the query
	stmt := fmt.Sprintf(
		SELECT_ROW,
		getDbTag(&AuthorsReviewers{}, "Id"),
		TABLE_AUTHORS,
		getDbTag(&AuthorsReviewers{}, "SubmissionId"),
	)
	// executes query
	rows, err := db.Query(stmt, submissionId)
	if err != nil {
		return nil, err
	}

	// builds the array
	var author string
	var authors []string
	for rows.Next() {
		// if there is an error returned by scanning the row, the error is returned
		// without the array
		if err := rows.Scan(&author); err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}
	return authors, nil
}

// Query the reviewers of a given submission from the database
//
// Params:
//	submissionId (int) : the id of the submission to get reviewers of
// Returns:
//	([]int) : of the reviewer's names
//	(error) : if something goes wrong during the query
func getSubmissionReviewers(submissionId int) ([]string, error) {
	// builds the query
	stmt := fmt.Sprintf(
		SELECT_ROW,
		getDbTag(&AuthorsReviewers{}, "Id"),
		TABLE_REVIEWERS,
		getDbTag(&AuthorsReviewers{}, "SubmissionId"),
	)
	// executes query
	rows, err := db.Query(stmt, submissionId)
	if err != nil {
		return nil, err
	}

	// builds the array
	var reviewer string
	var reviewers []string
	for rows.Next() {
		// if there is an error returned by scanning the row, the error is returned
		// without the array
		if err := rows.Scan(&reviewer); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewer)
	}
	return reviewers, nil
}

// Queries the database for file paths with the given submission ID
// (i.e. files in the submission)
//
// Params:
//	submissionId (int) : the id of the submission to get the files of
//Returns:
//	([]string) : of the file paths
//	(error) : if something goes wrong during the query
func getSubmissionFiles(submissionId int) ([]string, error) {
	// builds the query
	stmt := fmt.Sprintf(SELECT_ROW,
		getDbTag(&File{}, "Path"),
		TABLE_FILES,
		getDbTag(&File{}, "SubmissionId"),
	)
	// executes query
	rows, err := db.Query(stmt, submissionId)
	if err != nil {
		return nil, err
	}

	// builds the array
	var file string
	var files []string
	for rows.Next() {
		// if there is an error returned by scanning the row, the error is returned
		// without the array
		if err := rows.Scan(&file); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

// This function gets a submission's meta-data from its file in the filesystem
//
// Parameters:
// 	submissionId (int) : the unique id of the submission
// Returns:
//	(*CodeSubmissionData) : the submission's metadata if found
// 	(error) : if anything goes wrong while retrieving the metadata
func getSubmissionMetaData(submissionId int) (*CodeSubmissionData, error) {
	// gets the submission name from the database
	var submissionName string
	querySubmissionName := fmt.Sprintf(
		SELECT_ROW,
		getDbTag(&Submission{}, "Name"),
		TABLE_SUBMISSIONS,
		getDbTag(&Submission{}, "Id"),
	)
	row := db.QueryRow(querySubmissionName, submissionId)
	if err := row.Scan(&submissionName); err != nil {
		return nil, err
	}

	// reads the data file into a string
	dataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), DATA_DIR_NAME, submissionName+".json")
	dataString, err := ioutil.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	// marshalls the string of data into a struct
	submissionData := &CodeSubmissionData{}
	if err := json.Unmarshal(dataString, submissionData); err != nil {
		return nil, err
	}
	return submissionData, nil
}
