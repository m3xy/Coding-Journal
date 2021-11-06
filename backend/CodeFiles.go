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
- refine tags in db.go
- talk with alex about best practice for tags in structs (i.e. multi-table db tags)
	is table.column allowed for tags?
- factor out and generalize functionality for concatenating query results?
*/

package main

import (
	"encoding/json"
	"encoding/base64"
	"fmt"
	"net/http"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// file constants, includes
const (
	// TEMP: hard coded for testing
	FILESYSTEM_ROOT = "/home/ewp3/Documents/CS3099/project-code/testProjects/" // path to the root directory holding all project directories TEMP: maybe set with an env variable?
)

// structure to hold json data from data files
type CodeFileData struct {
	Comments []Comment `json:"comments"`
}

// // lightweight struct for listing project id/name pairs
// type ProjectsList struct {
// 	projects []
// 	id int `json:"id" db:"id"`
// 	name string `json:"name" db:"project_name"`
// }

// // structure to hold project data as read from json
// type ProjectData {

// }

/*
Function to populate File instance's fields by querying the SQL database

Params:
	file (*File) : a File structure instance to populate the fields of
Returns:
	error if something goes wrong while querying the DB
*/
func getFileData(file *File, fileID int) error {
	// queries the file path, project ID, and project name from the database
	stmt := fmt.Sprintf(SELECT_ROW_INNER_JOIN,
		getDbTag(&File{}, "path"),
		TABLE_FILES+", "+getDbTag(&File{}, "projectID")+", "+getDbTag(&File{}, "projectName"),
		TABLE_PROJECTS,
		TABLE_PROJECTS+"."+getDbTag(&Project{}, "id"),
		TABLE_FILES+"."+getDbTag(&File{}, "id"),
		TABLE_FILES+"."+getDbTag(&File{}, "id"),
	)
	// executes query (should only return 1 row via unique constraint on file ids)
	row := db.QueryRow(stmt, fileID)
	if err := row.Scan(file.path, file.projectID, file.projectName); err != nil {
		return err
	}

	// sets the file name in the object using the path
	file.name = filepath.Base(file.path)

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
		fmt.Sprint(file.projectID),
		file.projectName,
		file.path,
	)
	fileData, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// if no error occurred, encode and returns the file's content
	file.content = base64.StdEncoding.EncodeToString(fileData)
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
		fmt.Sprint(file.projectID), 
		".data", 
		file.projectName, 
		strings.TrimSuffix(file.path, filepath.Ext(file.path)) + ".json",
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
	file.comments = codeFileData.Comments
	return nil
}

/*
Router function to retrieve go code files from the filesystem. This function returns
the actual content of a code file

Params:
	path (String) : the path to the file to be retrieved, including the file name and
		extension
Returns:
	the contents of a file along with it's data
*/
func getFile(w http.ResponseWriter, r *http.Request) {
	// Set up writer response.
	w.Header().Set("Content-Type", "application/json")

	// creates the file object and gets its ID 
	file := &File{}
	err := json.NewDecoder(r.Body).Decode(file)
	if err != nil {
		// Bad request
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	// TEMP: add different responses for error handling here
	// populates all fields for the file except content and comments by querying the db
	if getFileData(file, int(file.id)) != nil {
		w.WriteHeader(http.StatusBadRequest)
	// gets the file contents, encodes it in Base64, and inserts it into the structure
	} else if getFileContent(file) != nil {
		w.WriteHeader(http.StatusBadRequest)
	// gets the file comments, and inserts them into the file structure
	} else if getFileComments(file) != nil {
		w.WriteHeader(http.StatusBadRequest)
	// writes JSON data for the file to the HTTP connection if no error has occured
	} else {
		response, err := json.Marshal(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(response)
	}
}

// /*
// Router function to get the names and id's of every project currently saved

// TEMP: impl some ordering based on views

// */
// func getAllProjects(w http.ResponseWriter, r *http.Request) {
// 	// set content type for return
// 	w.Header().Set("Content-Type", "application/json")

// 	// queries the database for the project ID and name pairs
// 	stmt := fmt.Sprintf(SELECT_ALL_ORDER_BY, TABLE_PROJECTS)
// 	rows, err := db.Query(stmt, )
// 	if err != nil {
// 		return nil, err
// 	}
// }

// create and return a new project struct
func newProject() *Project {
	return &Project{
		projectName: "",
		reviewers:    nil,
		authors:      nil,
	}
}

/*
Queries the authors of a given project from the database

Params:
	projectID (int) : the id of the project to get authors of
Returns:
	[]string of the author's names
	error if something goes wrong during the query
*/
func getProjectAuthors(projectID int) ([]string, error) {
	// builds the query
	stmt := fmt.Sprintf(SELECT_ROW_INNER_JOIN,
		getDbTag(&Credentials{}, "Fname")+", "+getDbTag(&Credentials{}, "Lname"),
		TABLE_USERS,
		TABLE_AUTHORS,
		TABLE_USERS+"."+getDbTag(&Credentials{}, "Id"),
		TABLE_AUTHORS+"."+"user_id",
		TABLE_AUTHORS+"."+"project_id",
	)
	// executes query
	rows, err := db.Query(stmt, projectID)
	if err != nil {
		return nil, err
	}

	// builds the array
	var fname, lname string
	var authors []string
	for rows.Next() {
		// if there is an error returned by scanning the row, the error is returned
		// without the array
		if err := rows.Scan(&fname, &lname); err != nil {
			return nil, err
		}
		authors = append(authors, fname+" "+lname)
	}
	return authors, nil
}

/*
Queries the reviewers of a given project from the database

Params:
	projectID (int) : the id of the project to get reviewers of
Returns:
	[]string of the reviewer's names
	error if something goes wrong during the query
*/
func getProjectReviewers(projectID int) ([]string, error) {
	// builds the query
	stmt := fmt.Sprintf(SELECT_ROW_INNER_JOIN,
		getDbTag(&Credentials{}, "Fname")+", "+getDbTag(&Credentials{}, "Lname"),
		TABLE_USERS,
		TABLE_REVIEWERS,
		TABLE_USERS+"."+getDbTag(&Credentials{}, "Id"),
		TABLE_REVIEWERS+"."+"user_id",
		TABLE_REVIEWERS+"."+"project_id",
	)
	// executes query
	rows, err := db.Query(stmt, projectID)
	if err != nil {
		return nil, err
	}

	// builds the array
	var fname, lname string
	var reviewers []string
	for rows.Next() {
		// if there is an error returned by scanning the row, the error is returned
		// without the array
		if err := rows.Scan(&fname, &lname); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, fname+" "+lname)
	}
	return reviewers, nil
}

/*
Queries the database for file paths with the given project ID
(i.e. files in the project)

Params:
	projectID (int) : the id of the project to get the files of
Returns:
	[]string of the file paths
	error if something goes wrong during the query
*/
func getProjectFiles(projectID int) ([]string, error) {
	// builds the query
	stmt := fmt.Sprintf(SELECT_ROW,
		getDbTag(&File{}, "path"),
		TABLE_FILES,
		getDbTag(&File{}, "id"),
	)
	// executes query
	rows, err := db.Query(stmt, projectID)
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

/*
Router function to retrieve code projects to be displayed on the web interface.
This function does not return the content of the code files, rather it'll return
the project name, ID and directory structure

TEMP: figure out what URL to have as endpoint here
*/
func getProject(w http.ResponseWriter, r *http.Request) {
	// Set up writer response.
	w.Header().Set("Content-Type", "application/json")

	// Decodes request into Project struct, which only sets the id field (in db.go)
	project := newProject()
	err := json.NewDecoder(r.Body).Decode(project)
	if err != nil {
		// Bad request
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// statement to query project name
	getProjectName := fmt.Sprintf(SELECT_ROW,
		getDbTag(&Project{}, "projectName"),
		TABLE_PROJECTS,
		getDbTag(&Project{}, "id"),
	)

	// executes query
	row := db.QueryRow(getProjectName, project.id)
	if row.Err() != nil {
		w.WriteHeader(http.StatusBadRequest)
		// TEMP: write something like "project does not exist"
	}
	if err := row.Scan(project.projectName); err != nil {
		// TEMP: do some error handling here
	}

	// gets project authors TEMP: add error handling
	project.authors, err = getProjectAuthors(project.id)
	project.reviewers, err = getProjectReviewers(project.id)
	project.filePaths, err = getProjectFiles(project.id)

	// writes JSON data for the project to the HTTP connection
	response, err := json.Marshal(project)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Write(response)
}
