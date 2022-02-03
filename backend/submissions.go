// =============================================================================
// submissions.go
// Authors: 190010425
// Created: November 18, 2021
//
// TODO: write functionality for popularity statistics and an ordering algorithm
// for suggested submissions
// TODO: make sure the creation date for the submissions is written to the db
// TODO: Change router function doc comments
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
// // 	"database/sql"
// // 	"encoding/json"
// 	"errors"
// 	"fmt"
// // 	"io/ioutil"
// // 	"log"
// // 	"net/http"
// 	"os"
// 	"path/filepath"
// // 	"strconv"

// 	"gorm.io/driver/mysql"
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/logger"
)

// ------
// Router Functions
// ------

// // Router function to get the names and id's of every submission of a given user
// // if no user id is given in the query parameters, return all valid submissions
// //
// // Response Codes:
// //	200 : if the action completed successfully
// // 	401 : if the proper security token was not given in the request
// //	500 : otherwise
// // Response Body:
// //	A JSON object of form: {...<submission id>:<submission name>...}
// func getAllSubmissions(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("[INFO] GetAllSubmissions request received from %v", r.RemoteAddr)
// 	// gets the userId from the URL
// 	var userId string
// 	params := r.URL.Query()
// 	userIds := params[getJsonTag(&Credentials{}, "Id")]
// 	if userIds == nil {
// 		userId = "*"
// 	} else {
// 		userId = userIds[0]
// 	}

// 	// set content type for return
// 	w.Header().Set("Content-Type", "application/json")
// 	// uses getUserSubmissions to get all user submissions by setting authorId = *
// 	submissions, err := getUserSubmissions(userId)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	// marshals and returns the map as JSON
// 	jsonString, err := json.Marshal(submissions)
// 	if err != nil {
// 		log.Printf("[ERROR] JSON formatting failed: %v", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	// writes json string
// 	log.Printf("[INFO] GetAllSubmission request from %v successful", r.RemoteAddr)
// 	w.Write(jsonString)
// }

// // Send submission data to the frontend for display. ID included for file
// // and comment queries.
// //
// // Response Codes:
// //	200 : if the submission exists and the request succeeded
// // 	401 : if the proper security token was not given in the request
// //	400 : if the request is invalid or badly formatted
// // 	500 : if something else goes wrong in the backend
// // Response Body:
// // 	submission: Object
// //		id: String
// //		name: String
// //		authors: Array of Author userIDs
// // 		reviewers: Array of Reviewer userIDs
// // 		files: Array of file path strings
// //		metadata: object
// //			abstract: String
// // 			reviews: array of objects
// // 				author: int
// //				time: datetime string
// //				base64Value: string
// //				replies: object (same as comments)
// func sendSubmission(w http.ResponseWriter, r *http.Request) {
// 	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
// 		return
// 	}
// 	if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		return
// 	}
// 	log.Printf("getSubmission request received from %v", r.RemoteAddr)

// 	w.Header().Set("Content-Type", "application/json")
// 	// gets the submission Id from the URL parameters
// 	params := r.URL.Query()
// 	submissionId, err := strconv.Atoi(params[getJsonTag(&Submission{}, "Id")][0])
// 	if err != nil {
// 		log.Printf("invalid id: %s\n", params[getJsonTag(&Submission{}, "Id")][0])
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	// gets the submission struct
// 	submission, err := getSubmission(submissionId)
// 	if err != nil {
// 		log.Printf("[ERROR] could not retrieve submission data properly: %v", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	// writes JSON data for the submission to the HTTP connection
// 	response, err := json.Marshal(submission)
// 	if err != nil {
// 		log.Printf("[ERROR] error formatting response: %v", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	log.Print("[INFO] success\n")
// 	w.Write(response)
// 	return
// }

// ------
// Helper Functions
// ------

// // Add submission to filesystem and database.
// // Note: the Files, Authors, and Reviewers fields should be empty here
// //
// // Params:
// //	submission (*Submission) : the submission to be added to the db 
// // 		(all fields but Id MUST be set)
// // Returns:
// //	(int) : the id of the added submission
// //	(error) : if the operation fails
// func addSubmission(submission *Submission) (int, error) {
// 	// error cases
// 	if submission == nil {
// 		return 0, errors.New("Submission cannot be nil")
// 	} else if submission.Name == "" {
// 		return 0, errors.New("Submission.Name must be set to a valid string")
// 	} else if submission.AuthorIDs == nil || len(submission.AuthorIDs) == 0 {
// 		return 0, errors.New("AuthorIDs array cannot be nil or length 0")
// 	// TODO: potentially make it so there must be at least 1 reviewer per submission
// 	} else if submission.ReviewerIDs == nil {
// 		return 0, errors.New("ReviewerIDs array cannot be nil")
// 	}

// 	// adds the submission to the db
// 	if err := gormDb.Transaction(func(tx *gorm.DB) error {
// 		// adds the submission to the db, automatically setting submission.ID
// 		if err := tx.Create(submission).Error; err != nil {
// 			return err
// 		} 

// 		// adds the author associations
// 		tx.Model(submission).Association("Authors").Append(submission.AuthorIDs)
// 	}); err != nil {
// 		return -1, err
// 	}


// 	// declares return values
// 	var submissionId int
// 	var err error

// 	// formats query to insert the submission into the db
// 	insertSubmission := fmt.Sprintf(
// 		INSERT_SUBMISSION,
// 		TABLE_SUBMISSIONS,
// 		getDbTag(&Submission{}, "Name"),
// 		getDbTag(&Submission{}, "License"),
// 	)

// 	// executes the query and gets the submission id
// 	row := db.QueryRow(insertSubmission, submission.Name, submission.License)
// 	// gets the id from the inserted submission
// 	if err = row.Scan(&submissionId); err != nil {
// 		return 0, err
// 	}

// 	// adds the authors and reviewers to their respective tables
// 	// (here we work with the assumption that author and reviewer arrays are very small)
// 	for _, authorId := range submission.AuthorIDs {
// 		if err = addAuthor(authorId, submissionId); err != nil {
// 			return 0, err
// 		}
// 	}
// 	for _, reviewerId := range submission.Reviewers {
// 		if err = addReviewer(reviewerId, submissionId); err != nil {
// 			return 0, err
// 		}
// 	}
// 	// adds the submission tags to the categories table
// 	for _, tag := range submission.Categories {
// 		if err = addTag(tag, submissionId); err != nil {
// 			return 0, err
// 		}
// 	}
// 	// builds an array of category structs to add to the db
	

// 	// creates the directories to hold the submission in the filesystem
// 	submissionPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), submission.Name)
// 	submissionDataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), DATA_DIR_NAME, submission.Name)
// 	if err = os.MkdirAll(submissionPath, DIR_PERMISSIONS); err != nil {
// 		return 0, err
// 	}
// 	if err = os.MkdirAll(submissionDataPath, DIR_PERMISSIONS); err != nil {
// 		return 0, err
// 	}

// 	// writes the submission metadata to it's corresponding file
// 	dataFile, err := os.OpenFile(
// 		submissionDataPath+".json", os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
// 	if err != nil {
// 		return 0, err
// 	}
// 	// Write submission metadata to the created file
// 	jsonString, err := json.Marshal(submission.MetaData)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if _, err = dataFile.Write([]byte(jsonString)); err != nil {
// 		return 0, err
// 	}
// 	dataFile.Close()

// 	// if the action was successful, the submission id of the submission struct is set and returned
// 	submission.Id = submissionId
// 	return submissionId, nil
// }

// // Add an author to the given submission provided the id given corresponds to a valid
// // user with publisher or publisher-reviewer permissions
// //
// // Params:
// //	authorId ([]string) : the global ids of the authors to add to the submission
// //	submissionId (int) : the id of the submission to be added to
// // Returns:
// //	(error) : an error if one occurs, nil otherwise
// func addAuthors(authorId []string, submissionId int) error {
// 	if submissionId < 0 {
// 		return errors.New(
// 			fmt.Sprintf("Submission IDs must be integers 0 or greater, not: %d", submissionId))
// 	}

// 	// checks that the author is a valid user with publisher or publisher-reviewer permissions
// 	var permissions int
// 	queryUserType := fmt.Sprintf(
// 		SELECT_ROW,
// 		getDbTag(&Credentials{}, "Usertype"),
// 		VIEW_PERMISSIONS,
// 		getDbTag(&IdMappings{}, "GlobalId"),
// 	)
// 	// executes the query, getting the user's permissions
// 	row := db.QueryRow(queryUserType, authorId)
// 	if err := row.Scan(&permissions); err != nil {
// 		return err
// 	}

// 	// checks permissions, and if they are correct, the author is added
// 	if permissions == USERTYPE_PUBLISHER || permissions == USERTYPE_REVIEWER_PUBLISHER {
// 		_, err := db.Query(fmt.Sprintf(INSERT_AUTHOR, TABLE_AUTHORS), submissionId, authorId)
// 		return err
// 	}
// 	return errors.New("User must be authorized as an author")
// }

// // Add a user to a submission as a reviewer
// //
// // Params:
// //	reviewerId (int) : the id of the reviewer to add to the submission
// //	submissionId (int) : the id of the submission to be added to
// // Returns:
// //	(error) : an error if one occurs, nil otherwise
// func addReviewer(reviewerId string, submissionId int) error {
// 	var err error
// 	if submissionId < 0 {
// 		return errors.New(fmt.Sprintf("Submission IDs must be integers 0 or greater, not: %d", submissionId))
// 	}

// 	// checks that the reviewer is a valid user with reviewer or publisher-reviewer permissions
// 	var permissions int
// 	queryUserType := fmt.Sprintf(
// 		SELECT_ROW,
// 		getDbTag(&Credentials{}, "Usertype"),
// 		VIEW_PERMISSIONS,
// 		getDbTag(&IdMappings{}, "GlobalId"),
// 	)
// 	// gets the user's permissions
// 	row := db.QueryRow(queryUserType, reviewerId)
// 	if err := row.Scan(&permissions); err != nil {
// 		return err
// 	}

// 	// checks permissions, and if they are correct, the reviewer is added
// 	if permissions == USERTYPE_REVIEWER || permissions == USERTYPE_REVIEWER_PUBLISHER {
// 		_, err = db.Query(fmt.Sprintf(INSERT_REVIEWER, TABLE_REVIEWERS), submissionId, reviewerId)
// 		return err
// 	}
// 	return errors.New("User must be authorized as a Reviewer")
// }

// // function to add tags for a given submission
// //
// // Parameters:
// // 	tags ([]string) : the tag to add to the submission
// // 	submissionId (int) : the unique Id of the submission to add tags to
// // Returns:
// // 	(error) : an error if one occurs, nil otherwise
// func addTags(tag string, submissionId int) (error) {
// 	if tag == "" {
// 		return fmt.Errorf("Tag cannot be empty") // TODO: figure out if this should err or just log
// 	}

// 	// makes sure the given tag does not already exist for the given project
// 	queryTag := fmt.Sprintf(
// 		SELECT_ROW,
// 		getDbTag(&Categories{}, "Tag"),
// 		TABLE_CATEGORIES,
// 		getDbTag(&Categories{}, "SubmissionId"),
// 	)
// 	// queries the db for the given submissions tags, and checks if they match the new tag
// 	rows, err := db.Query(queryTag, submissionId)
// 	var currTag string
// 	for rows.Next() {
// 		err = rows.Scan(&currTag)
// 		if err != nil {
// 			log.Printf("[ERROR] %v", err)
// 			return err
// 		}
// 		if currTag == tag {
// 			log.Print("[WARN] Tag already exists as specified")
// 			return errors.New("Tag already exists as specified")
// 		}
// 	}

// 	// adds the tag
// 	rows, err = db.Query(fmt.Sprintf(INSERT_CATEGORIES, TABLE_CATEGORIES), submissionId, tag)
// 	if rows != nil {
// 		rows.Close()
// 	}

// 	return err
// }

// // gets all authors which are written by a given user and returns them
// //
// // Params:
// // 	authorId (string) : the global id of the author as stored in the db
// // Return:
// // 	(map[int]string) : map of submission Ids to submission names
// // 	(error) : an error if something goes wrong, nil otherwise
// func getUserSubmissions(authorId string) (map[int]string, error) {
// 	// queries the database for the submission ID and name pairs
// 	stmt := fmt.Sprintf(
// 		SELECT_ROW,
// 		getDbTag(&AuthorsReviewers{}, "SubmissionId")+", "+
// 		getDbTag(&Submission{}, "Name"),
// 		VIEW_SUBMISSIONLIST, 
// 		"userId",
// 	)
// 	rows, err := db.Query(stmt, authorId)
// 	if err != nil {
// 		log.Printf("[WARN] User does exist: %s", authorId)
// 		return nil, err
// 	}

	// parses query result into { id : submission name } mappings
// 	var id int
// 	var submissionName string
// 	submissions := make(map[int]string)
// 	for rows.Next() {
// 		if err := rows.Scan(&id, &submissionName); err != nil {
// 			if err == sql.ErrNoRows {
// 				log.Printf("[WARN] Submission does not exist: %d", id)
// 				return nil, nil
// 			}
// 			log.Printf("[ERROR] SQL Error on submission retrieval.")
// 			return nil, err
// 		}
// 		submissions[id] = submissionName
// 	}
// 	return submissions, nil
// }

// // Get the submission struct corresponding to the id by querying the db and reading
// // the submissions meta-data file
// //
// // Parameters:
// // 	submissionId (int) : the submission's unique id
// // Returns:
// // 	(*Submission) : the data of the submission
// // 	(error) : an error if one occurs
// func getSubmission(submissionId int) (*Submission, error) {
// 	submission := &Submission{ Id: submissionId }

// 	// gets the submission name, creation date, and license
// 	querySubmissionData := fmt.Sprintf(
// 		SELECT_ROW,
// 		getDbTag(&Submission{}, "Name")+", "+
// 		getDbTag(&Submission{}, "CreationDate")+", "+
// 		getDbTag(&Submission{}, "License"),
// 		TABLE_SUBMISSIONS,
// 		getDbTag(&Submission{}, "Id"),
// 	)
// 	row := db.QueryRow(querySubmissionData, submissionId)
// 	err := row.Scan(&submission.Name,
// 		&submission.CreationDate, &submission.License)
// 	if err == sql.ErrNoRows {
// 		return nil, fmt.Errorf("No submission exists with ID: %d", submissionId)
// 	} else if err != nil {
// 		return nil, err
// 	}

// 	// gets the data which is not stored in the submissions table of the database
// 	submission.Authors, err = getSubmissionAuthors(submissionId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	submission.Reviewers, err = getSubmissionReviewers(submissionId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	submission.Categories, err = getSubmissionCategories(submissionId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	submission.FilePaths, err = getSubmissionFilePaths(submissionId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	submission.MetaData, err = getSubmissionMetaData(submissionId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return submission, nil
// }

// // Query the authors of a given submission from the database
// //
// // Params:
// //	submissionId (int) : the id of the submission to get authors of
// // Returns:
// //	([]string) : of the author's names
// //	(error) : if something goes wrong during the query
// func getSubmissionAuthors(submissionId int) ([]string, error) {
// 	// builds the query
// 	stmt := fmt.Sprintf(
// 		SELECT_ROW,
// 		getDbTag(&AuthorsReviewers{}, "Id"),
// 		TABLE_AUTHORS,
// 		getDbTag(&AuthorsReviewers{}, "SubmissionId"),
// 	)
// 	// executes query
// 	rows, err := db.Query(stmt, submissionId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// builds the array
// 	var author string
// 	var authors []string
// 	for rows.Next() {
// 		// if there is an error returned by scanning the row, the error is returned
// 		// without the array
// 		if err := rows.Scan(&author); err != nil {
// 			return nil, err
// 		}
// 		authors = append(authors, author)
// 	}
// 	return authors, nil
// }

// // Query the reviewers of a given submission from the database
// //
// // Params:
// //	submissionId (int) : the id of the submission to get reviewers of
// // Returns:
// //	([]string) : of the reviewer's Ids
// //	(error) : if something goes wrong during the query
// func getSubmissionReviewers(submissionId int) ([]string, error) {
// 	// builds the query
// 	stmt := fmt.Sprintf(
// 		SELECT_ROW,
// 		getDbTag(&AuthorsReviewers{}, "Id"),
// 		TABLE_REVIEWERS,
// 		getDbTag(&AuthorsReviewers{}, "SubmissionId"),
// 	)
// 	// executes query
// 	rows, err := db.Query(stmt, submissionId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// builds the array
// 	var reviewer string
// 	var reviewers []string
// 	for rows.Next() {
// 		// if there is an error returned by scanning the row, the error is returned
// 		// without the array
// 		if err := rows.Scan(&reviewer); err != nil {
// 			return nil, err
// 		}
// 		reviewers = append(reviewers, reviewer)
// 	}
// 	return reviewers, nil
// }

// // Queries the database for categories (tags) associated with the given
// // submission
// //
// // Parameters:
// // 	submissionId (int) : the submission to get the categories for
// // Returns:
// // 	([]string) : an array of the tags associated with the given submission
// // 	(error) : an error if one occurs while retrieving the tags
// func getSubmissionCategories(submissionId int) ([]string, error) {
// 	// gets categories of the submission 
// 	var currTag string
// 	categories := []string{}
// 	queryCategories := fmt.Sprintf(
// 		SELECT_ROW,
// 		getDbTag(&Categories{}, "Tag"),
// 		TABLE_CATEGORIES,
// 		getDbTag(&Categories{}, "SubmissionId"),
// 	)
// 	rows, err := db.Query(queryCategories, submissionId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// iterates over the results of the query, adding them to the array of tags
// 	for rows.Next() {
// 		if err = rows.Scan(&currTag); err != nil {
// 			return nil, err
// 		}
// 		categories = append(categories, currTag)
// 	}
// 	return categories, nil
// }

// // Queries the database for file paths with the given submission ID
// // (i.e. files in the submission)
// //
// // Params:
// //	submissionId (int) : the id of the submission to get the files of
// //Returns:
// //	([]string) : Array of file paths which are members of the given submission
// //	(error) : if something goes wrong during the query
// func getSubmissionFilePaths(submissionId int) ([]string, error) {
// 	// builds the query
// 	stmt := fmt.Sprintf(SELECT_ROW,
// 		getDbTag(&File{}, "Path"),
// 		TABLE_FILES,
// 		getDbTag(&File{}, "SubmissionId"),
// 	)
// 	// executes query
// 	rows, err := db.Query(stmt, submissionId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// builds the array
// 	var filePath string
// 	var files []string
// 	for rows.Next() {
// 		// returns error if one occurs
// 		if err := rows.Scan(&filePath); err != nil {
// 			return nil, err
// 		}
// 		files = append(files, filePath)
// 	}
// 	return files, nil
// }

// // This function gets a submission's meta-data from its file in the filesystem
// //
// // Parameters:
// // 	submissionId (int) : the unique id of the submission
// // Returns:
// //	(*SubmissionData) : the submission's metadata if found
// // 	(error) : if anything goes wrong while retrieving the metadata
// func getSubmissionMetaData(submissionId int) (*SubmissionData, error) {
// 	// gets the submission name from the database
// 	var submissionName string
// 	querySubmissionName := fmt.Sprintf(
// 		SELECT_ROW,
// 		getDbTag(&Submission{}, "Name"),
// 		TABLE_SUBMISSIONS,
// 		getDbTag(&Submission{}, "Id"),
// 	)
// 	row := db.QueryRow(querySubmissionName, submissionId)
// 	if err := row.Scan(&submissionName); err != nil {
// 		return nil, err
// 	}

// 	// reads the data file into a string
// 	dataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), DATA_DIR_NAME, submissionName+".json")
// 	dataString, err := ioutil.ReadFile(dataPath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// marshalls the string of data into a struct
// 	submissionData := &SubmissionData{}
// 	if err := json.Unmarshal(dataString, submissionData); err != nil {
// 		return nil, err
// 	}
// 	return submissionData, nil
// }

// // This function takes in a struct in the local submission format, and transforms
// // it into the supergroup compliant format
// //
// // Parameters:
// // 	submissionId (int) : the id of the submission to be converted to a supergroup-compliant
// // 		format
// // Returns:
// // 	(*SupergroupSubmission) : a supergroup compliant submission struct
// // 	(error) : an error if one occurs	
// func localToGlobal(submissionId int) (*SupergroupSubmission, error) {
// 	// gets the submission struct which submissionId refers to
// 	localSubmission, err := getSubmission(submissionId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// creates the Supergroup metadata struct
// 	supergroupData := &SupergroupSubmissionData{
// 		CreationDate: localSubmission.CreationDate,
// 		Categories: localSubmission.Categories,
// 		Abstract: localSubmission.MetaData.Abstract,
// 		License: localSubmission.License,
// 	}

// 	// builds a list of author names from the users table's first and last name fields
// 	queryAuthorNames := fmt.Sprintf(
// 		SELECT_ROW_INNER_JOIN,
// 		getDbTag(&Credentials{}, "Fname")+", "+getDbTag(&Credentials{}, "Lname"),
// 		VIEW_USER_INFO,
// 		TABLE_AUTHORS,
// 		VIEW_USER_INFO+"."+getDbTag(&IdMappings{}, "GlobalId"),
// 		TABLE_AUTHORS+"."+getDbTag(&AuthorsReviewers{}, "Id"),
// 		TABLE_AUTHORS+"."+getDbTag(&AuthorsReviewers{}, "SubmissionId"),
// 	)
// 	rows, err := db.Query(queryAuthorNames, submissionId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// iterates over each row to get the author names
// 	var fname string
// 	var lname string
// 	authorNames := []string{}
// 	for rows.Next() {
// 		if err = rows.Scan(&fname, &lname); err != nil {
// 			return nil, err
// 		}
// 		authorNames = append(authorNames, fname+" "+lname)
// 	}
// 	supergroupData.AuthorNames = authorNames

// 	// creates the list of file structs using the file paths and files.go
// 	var base64 string
// 	var supergroupFile *SupergroupFile
// 	supergroupFiles := []*SupergroupFile{}
// 	for _, path := range localSubmission.FilePaths {
// 		fullFilePath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(localSubmission.Id), localSubmission.Name, path)
// 		base64, err = getFileContent(fullFilePath)
// 		if err != nil {
// 			return nil, err
// 		}
// 		supergroupFile = &SupergroupFile{
// 			Name: filepath.Base(path),
// 			Base64Value: base64,
// 		}
// 		supergroupFiles = append(supergroupFiles, supergroupFile)
// 	}
	
// 	// creates the supergroup submission to return
// 	return &SupergroupSubmission{
// 		Name: localSubmission.Name,
// 		Files: supergroupFiles,
// 		MetaData: supergroupData,
// 	}, nil
// }

// // TODO : finish this function once the router function for exporting supergroup things is written
// // // Converts a supergroup compliant submission to the locally used format
// // //
// // // Parameters:
// // // 	globalSubmission (*SupergroupSubmission) : a supergroup compliant submission
// // // Returns:
// // // 	(*Submission) : the locally formatted version of the above submission
// // // 	(error) : an error if one occurs
// // func globalToLocal(globalSubmission *SupergroupSubmission) (*Submission, error) {
// // 	// creates the local metadata struct
// // 	localData := &SubmissionData{
// // 		Abstract: globalSubmission.MetaData.Abstract,
// // 		Reviews: nil,
// // 	}

// // 	// registers the submission

// // 	// creates the submission to return
// // 	localSubmission &Submission{
// // 		Name: globalSubmission.Name,
// // 		CreationDate: globalSubmission.MetaData.CreationDate,
// // 		License: globalSubmission.MetaData.License,
// // 		Reviewers: nil, // TODO: discuss with group, how do we make this non-nil
// // 		Authors: nil, // TODO: how do we do this??
// // 		FilePaths: ,
// // 		Categories: globalSubmission.MetaData.Categories,
// // 		MetaData: localData,
// // 	}, nil
// // }
