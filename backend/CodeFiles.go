/*
CodeFiles.go
author: 190010425
created: November 2, 2021

This file handles reading/writing code files along with their
data.

The directory structure for the filesystem is as follows:

Project ID (as stored in db Projects table)
	> <project_name>/ (as stored in the projects table)
		... (project directory structure)
	> .data/
		> project_data.json
		... (project directory structure)
notice that in the filesystem, the .data dir structure mirrors the
project, so that each file in the project can have a .json file storing
its data which is named in the same way as the source code (the only difference
being the extension)
*/

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
)

// file constants, includes
const (
	// TEMP: hard coded for testing
	FILESYSTEM_ROOT = "../filesystem/" // path to the root directory holding all project directories TEMP: maybe set with an env variable?
	DATA_DIR_NAME = ".data" // name of the hidden data dir to be put into the project directory structure

	// File Mode Constants
	DIR_PERMISSIONS  = 0755 // permissions for filesystem directories
	FILE_PERMISSIONS = 0644 // permissions for project files
)

// -----
// Upload router functions
// -----

// Upload lone code file to system. File is wrapped to dummy project with same name.
//
// Responses:
//	- 200 : if action completes successfully
func uploadSingleFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var request map[string]interface{}
	json.NewDecoder(r.Body).Decode(&request)

	// Parse data into local variables
	fileName := request[getJsonTag(&File{}, "Name")] // file name as a string
	fileAuthor := request["author"] // author's user Id
	fileContent := request[getJsonTag(&File{}, "Content")] // base64 encoding of file content

	// Put parsed values into a file object and a project object
	wrapperProject := &Project{
		Name: fileName.(string),
		Authors: []string{fileAuthor.(string)},
		FilePaths: []string{fileName.(string)},
	}
	file := &File{
		ProjectName: fmt.Sprint(fileName),
		Path: fileName.(string),
		Name: fileName.(string),
		Content: fileContent.(string),
	}

	// adds file to the db and filesystem
	projectId, err := addProject(wrapperProject)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	fileId, err := addFileTo(file, projectId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// formats json response
	response := map[string]string {
		"file": fmt.Sprint(fileId),
	}

	// writes fileId as response
	jsonString, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte(jsonString))
}

// -----
// Upload Helper Functions
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
	if project.Authors == nil {
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

// Add file into project, and store it to FS and DB.
// Note: Need valid project. No comments on file creation.
//
// Params:
//	file (*File) : the file to add to the db and filesystem (all fields but Id and ProjectId MUST be set)
//	projectId (int) : the id of the project which the added file is to be linked 
//		to as an unsigned integer
// Returns:
//	(int) : the id of the added file (0 if an error occurs)
//	(error) : if the operation fails
func addFileTo(file *File, projectId int) (int, error) {
	// declares return value variables 
	var fileId int
	var err error

	// formats SQL query to insert the file into the db
	insertFile := fmt.Sprintf(
		INSERT_FILE,
		TABLE_FILES,
		getDbTag(&File{}, "ProjectId"),
		getDbTag(&File{}, "Path"),
	)
	// executes the formatted query, returning the fileId 
	// (note that here SQL implicitly checks that the projectId exists in the projects table via Foreign key constraint)
	row := db.QueryRow(insertFile, projectId, file.Path)
	if row.Err() != nil {
		return 0, row.Err()
	}
	// gets the id from the just inserted file
	if err = row.Scan(&fileId); err != nil {
		return 0, err
	}

	// Add file to filesystem
	filePath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(projectId), file.ProjectName, file.Path)
	fileDataPath := filepath.Join(
		FILESYSTEM_ROOT,
		fmt.Sprint(projectId),
		DATA_DIR_NAME,
		file.ProjectName,
		strings.TrimSuffix(file.Path, filepath.Ext(file.Path)) + ".json",
	)
	// file paths without the file name (to create dirs if they don't exist yet)
	fileDirPath := filepath.Dir(filePath)
	fileDataDirPath := filepath.Dir(fileDataPath)

	// mkdir files's dir in case they don't yet exist
	if err = os.MkdirAll(fileDirPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	} else if err = os.MkdirAll(fileDataDirPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	}

	// Create and open file and it's corresponding data file
	codeFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
	if err != nil {
		return 0, err
	}
	dataFile, err := os.OpenFile(
		fileDataPath, os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
	if err != nil {
		return 0, err
	}

	// writes the file content
	if _, err = codeFile.Write([]byte(file.Content)); err != nil {
		return 0, err
	}
	// Write empty CodeFileData as json so comments and other data can be added later.
	jsonString, err := json.Marshal(&CodeFileData{})
	if err != nil {
		return 0, err
	}
	if _, err = dataFile.Write([]byte(jsonString)); err != nil {
		return 0, err
	}

	// closes files
	codeFile.Close()
	dataFile.Close()

	// Operation was successful ==> file Id set in file object and file returned.
	file.Id = fileId
	file.ProjectId = projectId
	return fileId, nil
}

// Add a user to a project as an author.
//
// Params:
//	authorId (int) : the id of the author to add to the project
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

	// executes the query, only returning one row
	row := db.QueryRow(queryUserType, authorId)
	if row.Err() != nil {
		return row.Err()
	}
	// gets the user's permissions 
	if err := row.Scan(&permissions); err != nil {
		return err
	}

	// checks permissions, and if they are correct, the author is added
	if permissions != USERTYPE_PUBLISHER && permissions != USERTYPE_REVIEWER_PUBLISHER {
		return errors.New("User must be authorized as Publisher" +
		"to be listed as project Author")
	} else {
		_, err := db.Query(
			fmt.Sprintf(INSERT_AUTHOR, TABLE_AUTHORS), projectId, authorId)
		return err
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

// Add a comment to a given file
//
// Params:
//	comment (*Comment) : The comment struct to add to the file
//	fileId (int) : the id of the file to add a comment to
// Returns:
//	(error) : an error if one occurs, nil otherwise
func addComment(comment *Comment, fileId int) error {
	var err error

	// queries the database to get the file path so that the file's data file can be found
	var projectId string
	var projectName string
	var filePath string
	// builds a query to get the file's name, project id, and it's project's name
	columns := fmt.Sprintf(
		"%s, %s, %s",
		TABLE_PROJECTS+"."+getDbTag(&Project{}, "Id"),
		getDbTag(&Project{}, "Name"),
		getDbTag(&File{}, "Path"),
	)
	queryPath := fmt.Sprintf(
		SELECT_ROW_INNER_JOIN,
		columns,
		TABLE_FILES,
		TABLE_PROJECTS,
		TABLE_FILES+"."+getDbTag(&File{}, "ProjectId"),
		TABLE_PROJECTS+"."+getDbTag(&Project{}, "Id"),
		TABLE_FILES+"."+getDbTag(&File{}, "Id"),
	)
	// executes the query
	row := db.QueryRow(queryPath, fileId)
	if row.Err() != nil {
		return row.Err()
	}

	// if the query was successful, build the file path
	if err = row.Scan(&projectId, &projectName, &filePath); err != nil {
		return err
	}
	dataFilePath := filepath.Join(
		FILESYSTEM_ROOT,
		fmt.Sprint(projectId),
		DATA_DIR_NAME,
		projectName,
		strings.TrimSuffix(filePath, filepath.Ext(filePath))+".json",
	)

	// reads the data file content and formats it into a CodeFileData struct
	data := &CodeFileData{}
	dataFileContent, err := ioutil.ReadFile(dataFilePath)
	if err = json.Unmarshal(dataFileContent, data); err != nil {
		return err
	}

	// adds the new comment and writes to the
	data.Comments = append(data.Comments, comment)
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// if everything has gone correctly, the new data is written to the file
	return ioutil.WriteFile(dataFilePath, dataBytes, FILE_PERMISSIONS)
}


// -----
// File functionality
// -----

// Retrieve code files from filesystem. Returns
// file content with comments and metadata.
// 
// Response Codes:
//	200 : File exists, getter success.
//	400 : otherwise
// Response Body:
// 		file: object
// 			fileName: string
// 			base64value: string
func getFile(w http.ResponseWriter, r *http.Request) {
	// Set up writer response.
	w.Header().Set("Content-Type", "application/json")

	// creates an empty project and gets the project id from the Get request header. If the header
	// does not contain an int value, return BadRequest header
	fileId, err := strconv.Atoi(r.Header.Get("file")) // TEMP: don't hard code this
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// TEMP: add different responses for error handling here
	// populates all fields for the file except content and comments by querying the db
	file := &File{Id: fileId}
	if err = getFileData(file, fileId); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	// gets the file contents and inserts it into the structure
	} else if err = getFileContent(file); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	// gets the file comments, and inserts them into the file structure
	} else if err = getFileComments(file); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// writes JSON data for the file to the HTTP connection if no error has occured
	response, err := json.Marshal(file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Write(response)
}

// Populate File instance's fields by querying the SQL database
//
// Params:
//	file (*File) : a File structure instance to populate the fields of
// Returns:
//	error if something goes wrong while querying the DB
func getFileData(file *File, fileId int) error {
	// queries the file path, project ID, and project name from the database
	queryColumns := fmt.Sprintf("%s, %s, %s",
		getDbTag(&File{}, "Path"),
		getDbTag(&File{}, "ProjectId"),
		getDbTag(&Project{}, "Name"),
	)
	stmt := fmt.Sprintf(SELECT_ROW_INNER_JOIN,
		queryColumns,
		TABLE_FILES,
		TABLE_PROJECTS,
		TABLE_FILES+"."+getDbTag(&File{}, "ProjectId"),
		TABLE_PROJECTS+"."+getDbTag(&Project{}, "Id"),
		TABLE_FILES+"."+getDbTag(&File{}, "Id"),
	)

	// executes query (should only return 1 row via unique constraint on file ids)
	row := db.QueryRow(stmt, fileId)
	if err := row.Scan(&file.Path, &file.ProjectId, &file.ProjectName); err != nil {
		return err
	}

	// sets the file name in the object using the path
	file.Name = filepath.Base(file.Path)

	// if no error has occurred, return nil
	return nil
}

// Get file content from filesystem. 
// Params:
// 	file (*File): pointer to File struct. All fields but content & comments must be set.
// Returns:
// 	(error) : if something goes wrong, nil otherwise
func getFileContent(file *File) error {
	// builds the path to the file and reads its content
	fullPath := filepath.Join(FILESYSTEM_ROOT,
		fmt.Sprint(file.ProjectId),
		file.ProjectName,
		file.Path,
	)
	// reads in the file's content
	fileData, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}
	// if no error occurred, assigns file.Content a value
	file.Content = string(fileData)
	return nil
}

// Get a file's comments from filesystem, and set Comments field
// 
// Params:
//	file (*File) : a pointer to a valid File struct. 
//	All fields must be set except for content and comments
//Returns:
//	(error) : if something goes wrong, nil otherwise
func getFileComments(file *File) error {
	// builds the path to the file and reads its content
	fullPath := filepath.Join(FILESYSTEM_ROOT,
		fmt.Sprint(file.ProjectId),
		DATA_DIR_NAME,
		file.ProjectName,
		strings.TrimSuffix(file.Path, filepath.Ext(file.Path))+".json",
	)
	jsonData, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}
	// fileData is parsed from json into the CodeFileData struct
	codeFileData := &CodeFileData{}
	err = json.Unmarshal(jsonData, codeFileData)
	if err != nil {
		return err
	}
	// if no error occurred, set the CodeFileData.comments field and return
	file.Comments = codeFileData.Comments
	return nil
}

// -----
// Project functionality
// -----

// Router function to get the names and id's of every project currently saved
//
// Response Codes:
//	200 : if the action completed successfully
//	400 : otherwise
// Response Body:
//	A JSON object of form: {...<project id>:<project name>...}
func getAllProjects(w http.ResponseWriter, r *http.Request) {
	// set content type for return
	w.Header().Set("Content-Type", "application/json")

	// queries the database for the project ID and name pairs
	columns := fmt.Sprintf("%s, %s",
		getDbTag(&Project{}, "Id"),
		getDbTag(&Project{}, "Name"),
	)
	stmt := fmt.Sprintf(SELECT_ALL_ORDER_BY, columns, TABLE_PROJECTS)
	rows, err := db.Query(stmt, getDbTag(&Project{}, "Name"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// parses query result into { id : project name } mappings
	var id int
	var projectName string
	projects := make(map[int]string)
	for rows.Next() {
		// if there is an error returned by scanning the row, the error is returned
		// without the array
		if err := rows.Scan(&id, &projectName); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		projects[id] = projectName
	}

	// marshals and returns the map as JSON
	jsonString, err := json.Marshal(projects)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// writes json string
	w.Write(jsonString)
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
	// Set up writer response.
	w.Header().Set("Content-Type", "application/json")

	// creates an empty project and gets the project id from the Get request header. If the header
	// does not contain an int value, return BadRequest header
	projectId, err := strconv.Atoi(r.Header.Get("project")) // TEMP: don't hard code this
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
		fmt.Println(row.Err())
		w.WriteHeader(http.StatusBadRequest)
	}
	// if no project name was returned for the given project id
	if err := row.Scan(&project.Name); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// gets project authors, reviewers, and file paths (as relpaths from the root of the project)
	project.Authors, err = getProjectAuthors(projectId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	project.Reviewers, err = getProjectReviewers(projectId)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
	}
	project.FilePaths, err = getProjectFiles(projectId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// writes JSON data for the project to the HTTP connection
	response, err := json.Marshal(project)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
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

// Queries the database for file paths with the given project ID
// (i.e. files in the project)
//
// Params:
//	projectId (int) : the id of the project to get the files of
//Returns:
//	([]int) : of the file paths
//	(error) : if something goes wrong during the query
func getProjectFiles(projectId int) ([]string, error) {
	// builds the query
	stmt := fmt.Sprintf(SELECT_ROW,
		getDbTag(&File{}, "Path"),
		TABLE_FILES,
		getDbTag(&File{}, "ProjectId"),
	)
	// executes query
	rows, err := db.Query(stmt, projectId)
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
