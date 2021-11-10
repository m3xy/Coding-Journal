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


NOTES:
- make function to query all project ids
- Talk with Manuel about enpoints for the project/files (i.e. w/ ids in URL)
- Maybe config path to dir holding projects with an environment variable?
- Maybe generalize the inner join query a bit more so that querying authors and reviewers is the same function?
- refine error handling
- Talk about what to store/return in structs for authors and reviewers (i.e. full name or ID)
- Do we want to store base64 text or raw source code in the backend filesystem? storing base64 is more efficient
*/

package main

import (
	"os"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"strconv"
	"errors"
)

// file constants, includes
const (
	// TEMP: hard coded for testing
	FILESYSTEM_ROOT = "/home/ewp3/Documents/CS3099/project-code/testProjects/" // path to the root directory holding all project directories TEMP: maybe set with an env variable?
	DATA_DIR_NAME = ".data" // name of the hidden data dir to be put into the project directory structure

	// File Mode Constants
	DIR_PERMISSIONS  = 0755 // permissions for filesystem directories
	FILE_PERMISSIONS = 0644 // permissions for project files
)

////////////////////////////////////////////////////////////////////////// HELPER FUNCTIONS FOR UPLOAD //////////////////////////////////////////////////////////////////////////

/*
helper function to add a project to the filesystem and database. Note that
the project id in this object should be set to -1 before this function is called
on it, because it will be set on addition to the db anyway. The file

Params:
	project (*Project) : the project to be added to the db (all fields but Id MUST be set)
Returns:
	(int) : the id of the added project
	(error) : if the operation fails
*/
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
		INSERT_PROJ, 
		TABLE_PROJECTS,
		getDbTag(&Project{}, "Name"),
	)

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

/*
helper function to add a file to the filesystem and database. Note that every
file must have a valid project to be attached to or else the adding will fail
Note that when a file is added, no commments or other data are present yet, and
hence it's data file will be empty.
This function should only be accessed from inside this file.

Params:
	file (*File) : the file to add to the db and filesystem (all fields but Id and ProjectId MUST be set)
	projectId (int) : the id of the project which the added file is to be linked 
		to as an unsigned integer
Returns:
	(int) : the id of the added file (0 if an error occurs)
	(error) : if the operation fails
*/
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

	// Adds the file to the filesystem
	filePath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(projectId), file.ProjectName, file.Path)
	fileDataPath := filepath.Join(
		FILESYSTEM_ROOT, 
		fmt.Sprint(projectId), 
		DATA_DIR_NAME, 
		file.ProjectName, 
		strings.TrimSuffix(file.Path, filepath.Ext(file.Path)) + ".json",
	)
	// file paths without the file name (for creating directories if they do not yet exist)
	fileDirPath := filepath.Dir(filePath)
	fileDataDirPath := filepath.Dir(fileDataPath)

	// creates the directories for the files in case they do not yet exist
	if err = os.MkdirAll(fileDirPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	} else if err = os.MkdirAll(fileDataDirPath, DIR_PERMISSIONS); err != nil {
		return 0, err
	}

	// creates and opens the file and it's corresponding data file
	codeFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
	if err != nil {
		return 0, err
	}
	dataFile, err := os.OpenFile(fileDataPath, os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
	if err != nil {
		return 0, err
	}

	// writes the file content
	if _, err = codeFile.Write([]byte(file.Content)); err != nil {
		return 0, err
	}
	// writes empty CodeFileData struct as json so that some comments and other data can be added later
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

	// if the operation was successful, the file id is set in the file object and the file is returned
	file.Id = fileId
	file.ProjectId = projectId
	return fileId, nil
}

/*
Adds a user to a project as an author

Params:
	authorId (int) : the id of the author to add to the project
	projectId (int) : the id of the project to be added to
Returns:
	(error) : an error if one occurs, nil otherwise
*/
func addAuthor(authorId int, projectId int) error {
	var err error
	if authorId < 0 {
		return errors.New(fmt.Sprintf("User IDs must be integers 0 or greater, not: %d", authorId))
	} else if projectId < 0 {
		return errors.New(fmt.Sprintf("Project IDs must be integers 0 or greater, not: %d", projectId))
	}

	// checks that the author is a valid user with publisher or publisher-reviewer permissions
	var permissions int
	queryUserType := fmt.Sprintf(
		SELECT_ROW,
		getDbTag(&Credentials{}, "Usertype"),
		TABLE_USERS,
		getDbTag(&Credentials{}, "Id"),
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
		return errors.New("User must be authorized as a Publisher to be listed as a project Author")
	} else {
		_, err = db.Query(fmt.Sprintf(INSERT_AUTHOR, TABLE_AUTHORS), projectId, authorId)
		return err
	}
}

/*
Adds a user to a project as a reviewer

Params:
	reviewerId (int) : the id of the reviewer to add to the project
	projectId (int) : the id of the project to be added to
Returns:
	(error) : an error if one occurs, nil otherwise
*/
func addReviewer(reviewerId int, projectId int) error {
	var err error
	if reviewerId < 0 {
		return errors.New(fmt.Sprintf("Reviewer IDs must be integers 0 or greater, not: %d", reviewerId))
	} else if projectId < 0 {
		return errors.New(fmt.Sprintf("Project IDs must be integers 0 or greater, not: %d", projectId))
	}

	// checks that the reviewer is a valid user with reviewer or publisher-reviewer permissions
	var permissions int
	queryUserType := fmt.Sprintf(
		SELECT_ROW,
		getDbTag(&Credentials{}, "Usertype"),
		TABLE_USERS,
		getDbTag(&Credentials{}, "Id"),
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

/*
Adds a comment to a given file

Params:
	comment (*Comment) : The comment struct to add to the file
	fileId (int) : the id of the file to add a comment to
Returns:
	(error) : an error if one occurs, nil otherwise
*/ 
func addComment(comment *Comment, fileId int) error {
	var err error

	// queries the database to get the file path so that the file's data file can be found
	var projectId string 
	var projectName string
	var filePath string
	// builds a query to get the file's name, project id, and it's project's name
	columns := fmt.Sprintf(
		"%s, %s, %s", 
		getDbTag(&Project{}, "Id"), 
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
	data.Comments = append(data.Comments, *comment)
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err 
	}

	// if everything has gone correctly, the new data is written to the file
	return ioutil.WriteFile(dataFilePath, dataBytes, FILE_PERMISSIONS)
}


////////////////////////////////////////////////////////////////////////////// FILES FUNCTIONALITY //////////////////////////////////////////////////////////////////////////////

/*
Router function to retrieve go code files from the filesystem. This function returns
the actual content of a code file along with comments and some data

Response Codes:
	200 : if the file exists, and no errors occurred while retrieving it
	400 : otherwise
Response Body:
	A Json marshalled File struct (in db.go) with content being a Base64 string
*/
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
	// gets the file contents, encodes it in Base64, and inserts it into the structure
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

/*
Function to populate File instance's fields by querying the SQL database

Params:
	file (*File) : a File structure instance to populate the fields of
Returns:
	error if something goes wrong while querying the DB
*/
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

/*
Helper function to get a file's content from the filesystem, and return its Base64 encoding

Params:
	file (*File) : a pointer to a valid File struct. All fields must be set except for content and comments
Returns:
	*string of the file's contents encoded in Base64 (uses pointer so that the file content is never
		moved or copied)
*/
func getFileContent(file *File) error {
	// builds the path to the file and reads its content
	fullPath := filepath.Join(FILESYSTEM_ROOT,
		fmt.Sprint(file.ProjectId),
		file.ProjectName,
		file.Path,
	)
	fileData, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// if no error occurred, encode and returns the file's content
	file.Content = base64.StdEncoding.EncodeToString(fileData)
	return nil
}

/*
Helper function to get a file's content from the filesystem, and return its Base64 encoding

Params:
	file (*File) : a pointer to a valid File struct. All fields must be set except for content and comments
Returns:
	*string of the file's contents encoded in Base64 (uses pointer so that the file content is never
		moved or copied)
*/
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

////////////////////////////////////////////////////////// PROJECTS FUNCTIONALITY ///////////////////////////////////////////////////////////

/*
Router function to get the names and id's of every project currently saved

Response Codes:
	200 : if the action completed successfully
	400 : otherwise
Response Body:
	A JSON object of form: {...<project id>:<project name>...}
*/
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

/*
Router function to retrieve a code project to be displayed on the web interface.
This function does not return the content of the code files, rather it'll return
the project name, ID and directory structure, so that the code files can be individually
queried

TEMP: figure out what URL to have as endpoint here

Response Codes:
	200 : if the project exists and the request succeeded
	400 : otherwise
Response Body:
	A marshalled Project struct (contained in db.go) 
*/
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
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Write(response)
}

/*
Queries the authors of a given project from the database

Params:
	projectId (int) : the id of the project to get authors of
Returns:
	[]string of the author's names
	error if something goes wrong during the query
*/
func getProjectAuthors(projectId int) ([]int, error) {
	// builds the query
	stmt := fmt.Sprintf(
		SELECT_ROW,
		"userId", 
		TABLE_AUTHORS,
		"projectId",
	)
	// executes query
	rows, err := db.Query(stmt, projectId)
	if err != nil {
		return nil, err
	}

	// builds the array
	var author int
	var authors []int
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

/*
Queries the reviewers of a given project from the database

Params:
	projectId (int) : the id of the project to get reviewers of
Returns:
	([]int) : of the reviewer's names
	(error) : if something goes wrong during the query
*/
func getProjectReviewers(projectId int) ([]int, error) {
	// builds the query
	stmt := fmt.Sprintf(
		SELECT_ROW,
		"userId", 
		TABLE_REVIEWERS,
		"projectId",
	)
	// executes query
	rows, err := db.Query(stmt, projectId)
	if err != nil {
		return nil, err
	}

	// builds the array
	var reviewer int
	var reviewers []int
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

/*
Queries the database for file paths with the given project ID
(i.e. files in the project)

Params:
	projectId (int) : the id of the project to get the files of
Returns:
	([]int) : of the file paths
	(error) : if something goes wrong during the query
*/
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
