/*
submissions.go
author: 190010425
created: November 18, 2021

This file handles the reading/writing of all project submissions
*/

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	// "io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	// "time"
)

// ----- 
// Add submissions functionality
// ----- 

// Add project to filesystem and database.
// Note: project ID is set by this function.
// Params:
//	project (*Project) : the project to be added to the db (all fields but Id MUST be set)
// Returns:
//	(int) : the id of the added project
//	(error) : if the operation fails
func addProject(project *Project) (int, error) {
	// error cases
	if project == nil {
		return 0, errors.New("Project cannot be nil")
	} else if project.Name == "" {
		return 0, errors.New("Project.Name must be set to a valid string")
	} else if project.Authors == nil {
		return 0, errors.New("Authors array cannot be nil")
	} else if project.Reviewers == nil {
		return 0, errors.New("Reviewers array cannot be nil")
	}

	// declares return values
	var projectId int
	var err error

	// formats query to insert the project into the db
	insertProject := fmt.Sprintf(
		INSERT_PROJ, TABLE_PROJECTS,
		getDbTag(&Project{}, "Name"))

	// executes the query and gets the project id
	row := db.QueryRow(insertProject, project.Name)
	if row.Err() != nil {
		return 0, row.Err()
	}
	// gets the id from the inserted project
	if err = row.Scan(&projectId); err != nil {
		return 0, err
	}

	// adds the authors and reviewers to their respective tables
	// (here we work with the assumption that author and reviewer arrays are very small)
	for _, authorId := range project.Authors {
		if err = addAuthor(authorId, projectId); err != nil {
			return 0, err
		}
	}
	for _, reviewerId := range project.Reviewers {
		if err = addReviewer(reviewerId, projectId); err != nil {
			return 0, err
		}
	}

	// adds a project to the mock filesystem
	projectPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(projectId), project.Name)
	projectDataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(projectId), DATA_DIR_NAME, project.Name)
	if err = os.MkdirAll(projectPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	}
	if err = os.MkdirAll(projectDataPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	}

	// if the action was successful, the project id of the project struct is set and returned
	project.Id = projectId
	return projectId, nil
}

// Add an author to the given project provided the id given corresponds to a valid
// user with publisher or publisher-reviewer permissions
//
// Params:
//	authorId (int) : the global id of the author to add to the project
//	projectId (int) : the id of the project to be added to
// Returns:
//	(error) : an error if one occurs, nil otherwise
func addAuthor(authorId string, projectId int) error {
	if projectId < 0 {
		return errors.New(
			fmt.Sprintf("Project IDs must be integers 0 or greater, not: %d", projectId))
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
			fmt.Sprintf(INSERT_AUTHOR, TABLE_AUTHORS), projectId, authorId)
		return err
	} else {
		return errors.New("User must be authorized as Publisher " +
			"to be listed as project Author" + fmt.Sprint(permissions))
	}
}

// Add a user to a project as a reviewer
//
// Params:
//	reviewerId (int) : the id of the reviewer to add to the project
//	projectId (int) : the id of the project to be added to
// Returns:
//	(error) : an error if one occurs, nil otherwise
func addReviewer(reviewerId string, projectId int) error {
	var err error
	if projectId < 0 {
		return errors.New(fmt.Sprintf("Project IDs must be integers 0 or greater, not: %d", projectId))
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
		_, err = db.Query(fmt.Sprintf(INSERT_REVIEWER, TABLE_REVIEWERS), projectId, reviewerId)
		return err
	}
}


// -----
// Retrieve Submission Functionality
// ----- 

// Router function to get the names and id's of every project currently saved
//
// Response Codes:
//	200 : if the action completed successfully
//	400 : otherwise
// Response Body:
//	A JSON object of form: {...<project id>:<project name>...}
func getAllProjects(w http.ResponseWriter, r *http.Request) {
	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
		return
	}
	if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// set content type for return
	w.Header().Set("Content-Type", "application/json")
	// uses getUserProjects to get all user projects by setting authorId = *
	projects, err := getUserProjects("*")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// marshals and returns the map as JSON
	jsonString, err := json.Marshal(projects)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// writes json string
	w.Write(jsonString)
}

// gets all authors which are written by a given user and returns them
//
// Params:
// 	authorId (string) : the global id of the author as stored in the db
// Return:
// 	(map[int]string) : map of project Ids to project names
// 	(error) : an error if something goes wrong, nil otherwise
func getUserProjects(authorId string) (map[int]string, error) {
	// queries the database for the project ID and name pairs
	columns := fmt.Sprintf("%s, %s",
		getDbTag(&AuthorsReviewers{}, "ProjectId"),
		getDbTag(&Project{}, "Name"),
	)
	stmt := fmt.Sprintf(SELECT_ROW, columns, VIEW_PROJECTLIST, "userId")
	rows, err := db.Query(stmt, authorId)
	if err != nil {
		return nil, err
	}

	// parses query result into { id : project name } mappings
	var id int
	var projectName string
	projects := make(map[int]string)
	for rows.Next() {
		// if there is an error returned by scanning the row, the error is returned
		// without the array
		if err := rows.Scan(&id, &projectName); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		projects[id] = projectName
	}
	return projects, nil
}

// Get project for display on frontend. ID included for file and comment queries.
//
// TODO figure out what URL to have as endpoint here
//
// Response Codes:
//	200 : if the project exists and the request succeeded
//	400 : otherwise
// Response Body:
//	A marshalled Project struct (contained in db.go)
func getProject(w http.ResponseWriter, r *http.Request) {
	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
		return
	}
	if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Set up writer response.
	w.Header().Set("Content-Type", "application/json")

	// gets the file path and project Id from the request body
	var request map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	projectId := int(request[getJsonTag(&Project{}, "Id")].(float64))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// creates a project with the given ID
	project := &Project{Id: projectId}

	// statement to query project name
	getProjectName := fmt.Sprintf(SELECT_ROW,
		getDbTag(&Project{}, "Name"),
		TABLE_PROJECTS,
		getDbTag(&Project{}, "Id"),
	)

	// executes query
	row := db.QueryRow(getProjectName, projectId)
	if row.Err() != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// if no project name was returned for the given project id
	if err := row.Scan(&project.Name); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// gets project authors, reviewers, and file paths (as relpaths from the root of the project)
	project.Authors, err = getProjectAuthors(projectId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	project.Reviewers, err = getProjectReviewers(projectId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	project.FilePaths, err = getProjectFiles(projectId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// writes JSON data for the project to the HTTP connection
	response, err := json.Marshal(project)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Write(response)
}


// Query the authors of a given project from the database
//
// Params:
//	projectId (int) : the id of the project to get authors of
// Returns:
//	[]string of the author's names
//	error if something goes wrong during the query
func getProjectAuthors(projectId int) ([]string, error) {
	// builds the query
	stmt := fmt.Sprintf(
		SELECT_ROW,
		getDbTag(&AuthorsReviewers{}, "Id"),
		TABLE_AUTHORS,
		getDbTag(&AuthorsReviewers{}, "ProjectId"),
	)
	// executes query
	rows, err := db.Query(stmt, projectId)
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

// Query the reviewers of a given project from the database
//
// Params:
//	projectId (int) : the id of the project to get reviewers of
// Returns:
//	([]int) : of the reviewer's names
//	(error) : if something goes wrong during the query
func getProjectReviewers(projectId int) ([]string, error) {
	// builds the query
	stmt := fmt.Sprintf(
		SELECT_ROW,
		getDbTag(&AuthorsReviewers{}, "Id"),
		TABLE_REVIEWERS,
		getDbTag(&AuthorsReviewers{}, "ProjectId"),
	)
	// executes query
	rows, err := db.Query(stmt, projectId)
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

// -----
// Supergroup migration
// -----

// Router function to import Journal submissions (projects) from other journals
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
	// set content type for return
	w.Header().Set("Content-Type", "application/json")

	// parses the data into a structure
	var submission *SupergroupSubmission
	err := json.NewDecoder(r.Body).Decode(submission)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// moves the data from the structure into a locally compliant format
	// don't receive user email, so it creates a fake one
	authorEmail := fmt.Sprintf("%s@email.com",
		strings.Replace(submission.Metadata.AuthorName, " ", "_", 4))
	author := &Credentials{
		Fname:    strings.Split(submission.Metadata.AuthorName, " ")[0],
		Lname:    strings.Split(submission.Metadata.AuthorName, " ")[1],
		Email:    authorEmail,
		Pw:       "password", // defaults to password here as we have no way of identifying users
		Usertype: USERTYPE_PUBLISHER,
	}
	authorId, err := registerUser(author)
	// formats the data in a project
	project := &Project{
		Name:      strings.Replace(submission.Name, " ", "_", 10), // default is 10 spaces to replace
		Reviewers: []string{},
		Authors:   []string{authorId},
		FilePaths: []string{},
	}

	// adds the project
	projectId, err := addProject(project)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	project.Id = projectId

	// adds the file
	for _, submissionFile := range submission.Files {
		file := &File{
			ProjectId:   projectId,
			ProjectName: project.Name,
			Name:        submissionFile.Name,
			Path:        submissionFile.Name,
			Content:     submissionFile.Content,
		}
		_, err = addFileTo(file, projectId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	// writes status OK if nothing goes wrong
	w.WriteHeader(http.StatusOK)
}

