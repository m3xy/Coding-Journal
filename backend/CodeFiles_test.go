/*
CodeFiles_test.go
author: 190010425
created: November 2, 2021

Test file for the CodeFiles module.

Note that the tests are written dependency wise from top to bottom. This means
that if a test breaks, then most of the tests below it will also break. Hence if
a test breaks, fix the top one first and then re-run
*/

package main

import (
	// "context"
	// "encoding/json"
	// "encoding/base64"
	"fmt"
	// "net/http"
	"os"
	// "path/filepath"
	// "strings"
	"testing"
	"errors"
	"reflect"

	// "github.com/gorilla/mux"
)

const (
	// constants for filesystem
	// TEST_DB = "testdb" // TEMP: declared in authentication_test.go

	// BE VERY CAREFUL WITH THIS PATH!! IT GETS RECURSIVELY REMOVED!!
	TEST_FILES_DIR = "/home/ewp3/Documents/CS3099/project-code/testProjects/" // environment variable set to this value

	TEST_URL = "http://localhost"
	TEST_SERVER_PORT = "3333"
)

// Constants for testing
var testProjects []*Project = []*Project{
	{Id: -1, Name: "testProject1", Reviewers: []int{},
		Authors: []int{}, FilePaths: []string{"testFile1.txt"}},
	{Id: -1, Name: "testProject2", Reviewers: []int{},
		Authors: []int{}, FilePaths: []string{"testFile2.txt"}},
}
var testFiles []*File = []*File{
	{Id: -1, ProjectId: -1, ProjectName: "testProject1", Path: "testFile1.txt",
		Name: "testFile1.txt", Content: "hello world", Comments: nil},
	{Id: -1, ProjectId: -1, ProjectName: "testProject2", Path: "testFile2.txt",
		Name: "testFile2.txt", Content: "hello world", Comments: nil},
}
var testAuthors []*Credentials = []*Credentials {
	{Email: "test@test.com", Pw: "123456aB$", Fname: "test",
		Lname: "test", PhoneNumber: "0574349206", Usertype: USERTYPE_PUBLISHER},
	{Email: "john.doe@test.com", Pw:"dlbjDs2!", Fname: "John",
		Lname: "Doe", Organization: "TestOrg", Usertype: USERTYPE_USER},
	{Email: "jane.doe@test.net", Pw:"dlbjDs2!", Fname: "Jane",
		Lname: "Doe", Usertype: USERTYPE_REVIEWER},
	{Email: "adam.doe@test.net", Pw:"dlbjDs2!", Fname: "Adam",
		Lname: "Doe", Usertype: USERTYPE_REVIEWER_PUBLISHER},
}
var testReviewers []*Credentials = []*Credentials {
	{Email: "dave@test.com", Pw: "123456aB$", Fname: "dave",
		Lname: "smith", PhoneNumber: "0574349206", Usertype: USERTYPE_REVIEWER},
	{Email: "Geoff@test.com", Pw:"dlbjDs2!", Fname: "Geoff",
		Lname: "Williams", Organization: "TestOrg", Usertype: USERTYPE_USER},
	{Email: "jane.doe@test.net", Pw:"dlbjDs2!", Fname: "Jane",
		Lname: "Doe", Usertype: USERTYPE_PUBLISHER},
	{Email: "adam.doe@test.net", Pw:"dlbjDs2!", Fname: "Adam",
		Lname: "Doe", Usertype: USERTYPE_REVIEWER_PUBLISHER},
}
var testComments []Comment = []Comment{}
var testFileData []*CodeFileData = []*CodeFileData{
	{Comments: testComments},
}

/*
initializes and clears the test database and filesystem, deleting and pre-existing entries
*/
func initTestEnvironment() error {
	// initializes the database
	dbInit(user, password, protocol, host, port, TEST_DB)

	// empties all db tables
	tablesToClear := []string{TABLE_USERS, TABLE_AUTHORS, TABLE_FILES, TABLE_PROJECTS, TABLE_REVIEWERS}
	for _, table := range tablesToClear {
		stmt := fmt.Sprintf(DELETE_ALL_ROWS, table)
		_, err := db.Query(stmt)
		if err != nil {
			return err
		}
	}
	// initializes the test filesystem
	if err := os.Mkdir(TEST_FILES_DIR, DIR_PERMISSIONS); err != nil {
		return err
	}

	return nil
}

/*
Function to remove the test filesystem and clear the database for the next test
*/
func clearTestEnvironment() error {
	// deletes the test filesystem
	if err := os.RemoveAll(TEST_FILES_DIR); err != nil {
		return err
	}
	// TEMP: issue with hanging on these commands?
	// // destroys the db
	// tablesToClear := []string{TABLE_FILES, TABLE_PROJECTS, TABLE_AUTHORS, TABLE_REVIEWERS, TABLE_USERS}
	// for _, table := range tablesToClear {
	// 	fmt.Println(table)
	// 	stmt := fmt.Sprintf(DELETE_ALL_ROWS, table)
	// 	_, err := db.Query(stmt)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// closes the connection to the db
	dbCloseConnection()
	return nil
}

/*
  Tests the functionality to create projects in the database and filesystem from a 
  Project struct
*/
func TestCreateProject(t *testing.T) {
	var err error

	// tests basic functionality with a valid test project. re-used in many tests
	testAddProject := func(testProject *Project) error {
		projectId, err := addProject(testProject)

		// simple error cases
		if err != nil {
			return err
		} else if projectId < 0 {
			return errors.New(fmt.Sprintf("Invalid Project ID returned: %d", projectId))
		}

		// checks manually that the project was added correctly
		var projectName string
		authors := []int{}
		reviewers := []int{}
		// builds SQL Queries for testing the added values
		queryProjectName := fmt.Sprintf(
			SELECT_ROW,
			getDbTag(&Project{}, "Name"),
			TABLE_PROJECTS,
			getDbTag(&Project{}, "Id"),
		)
		queryAuthors := fmt.Sprintf(
			SELECT_ROW,
			"userId",
			TABLE_AUTHORS,
			"projectId",
		)
		queryReviewers := fmt.Sprintf(
			SELECT_ROW,
			"userId",
			TABLE_REVIEWERS,
			"projectId",
		)

		// tests that the project name was added correctly
		row := db.QueryRow(queryProjectName, projectId)
		if row.Err() != nil {
			return errors.New(fmt.Sprintf("error querying project name: %v", row.Err()))
		} else if err = row.Scan(&projectName); err != nil {
			return err
		} else if testProject.Name != projectName {
			return errors.New(fmt.Sprintf("project names do not match. Entered: %s Gotten Back: %s", testProject.Name, projectName))
		}

		// tests that the authors were added correctly
		rows, err := db.Query(queryAuthors, projectId)
		if err != nil {
			return errors.New(fmt.Sprintf("error querying project Authors: %v", err))
		}
		var author int
		for rows.Next() {
			rows.Scan(&author)
			authors = append(authors, author)
		}
		if !(reflect.DeepEqual(testProject.Authors, authors)) {
			return errors.New("authors arrays do not match")
		}

		// tests that the reviewers were added correctly
		rows, err = db.Query(queryReviewers, projectId)
		if err != nil {
			return errors.New(fmt.Sprintf("error querying project Reviewers: %v", err))
		}
		var reviewer int
		for rows.Next() {
			rows.Scan(&reviewer)
			reviewers = append(reviewers, reviewer)
		}
		if !(reflect.DeepEqual(testProject.Reviewers, reviewers)) {
			return errors.New("reviewers arrays do not match")
		}
		return nil
	}

	// wrapper to init and teardown test environment to test adding a single project
	testAddSingleProj := func(project *Project) error {
		initTestEnvironment()
		// adds the project and tests that it was added properly
		if err = testAddProject(project); err != nil {
			return err // error already formatted here
		}
		// tears down the test environment
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while tearing down db: %v", err))
		}
		return nil
	}

	// test to add n projects in a row
	testAddNProjects := func(projects []*Project) error {
		initTestEnvironment()
		for _, project := range projects {
			if err = testAddProject(project); err != nil {
				return err
			}
		}
		// tears down the test environment
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while tearing down db: %v", err))
		}
		return nil
	}

	// runs tests
	if err = testAddSingleProj(testProjects[0]); err != nil {
		t.Errorf("testAddSingleProj failed for testProjects[0]: %v", err)
	} else if err = testAddNProjects(testProjects[0:2]); err != nil {
		t.Errorf("testAddNProjects failed for testProjects[0:2]: %v", err)
	}
}

/*
 This function tests adding authors to a given project.

 Test Depends on:
	- TestCreateProject()
	- TestRegisterUser() (in authentication.go)
*/
func TestAddAuthor(t *testing.T) {
	var err error

	// test to add a single valid author
	testSingleValidAuthor := func (author *Credentials) error {
		initTestEnvironment()

		// declares test variables
		var projectId int
		var authorId int
		testProject := testProjects[0]

		// adds a valid project and user to the db and filesystem so that an author can be added
		projectId, err = addProject(testProject)
		if err != nil {
			return err
		}
		err = registerUser(author)
		if err != nil {
			return errors.New(fmt.Sprintf("error registering the author: %v", err))
		}

		// TEMP: workaround because register user doesn't return an id
		// gets the just added user's user Id to add them as an author
		queryUserId := fmt.Sprintf(
			SELECT_ROW,
			getDbTag(&Credentials{}, "Id"),
			TABLE_USERS,
			getDbTag(&Credentials{}, "Email"),
		)
		// queries the userId from the db
		row := db.QueryRow(queryUserId, author.Email)
		if row.Err() != nil {
			return errors.New(fmt.Sprintf("error getting author Id from the db: %v", row.Err()))
		} else if err = row.Scan(&authorId); err != nil {
			return errors.New(fmt.Sprintf("error getting author Id from the db: %v", err))
		}
		
		// adds the author to the database
		if err = addAuthor(authorId, projectId); err != nil {
			return errors.New(fmt.Sprintf("error adding the author to the db: %v", err))
		}



		// clears the test environmtent and returns nil because the test has passed
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while tearing down db: %v", err))
		}
		return nil 
	}

	// runs tests
	if err = testSingleValidAuthor(testAuthors[0]); err != nil {
		t.Errorf("testSingleValidAuthor failed for testAuthors[0]: %v", err)
	}
}

/*
 This function tests adding authors to a given project. Note that this test uses the
 addProject functionality, and as such will fail if it fails

 Test Depends on:
	- TestCreateProject()
	- TestRegisterUser() (in authentication.go)
*/
func TestAddReviewers(t *testing.T) {
	
}

/*
 This function tests adding authors to a given project. Note that this test uses the
 addProject functionality, and as such will fail if it fails

 Test Depends on:
	- TestCreateProject()
*/
func TestAddFile(t *testing.T) {
	
}


// /*
// This function takes in a project data type and adds it to the test database
// and filesystem. This function sets the project id upon insertion into the db

// Params:
// 	project (*Project) : a project object to be inserted into the db
// Returns:
// 	(int) : the project id as inserted into the db if the operation is successful (-1 if not)
// 	(error) : an error if one occurs
// */
// func addProject(project *Project) (int, error) {
// 	// inserts data into the db
// 	var projectId int
// 	var err error
// 	insertProject := fmt.Sprintf(INSERT_PROJ, 
// 		TABLE_PROJECTS,
// 		getDbTag(&Project{}, "Name"))
// 	row := db.QueryRow(insertProject, project.Name)
// 	if row.Err() != nil {
// 		return -1, row.Err()
// 	}
// 	// gets the id from the just inserted project
// 	if err = row.Scan(&projectId); err != nil {
// 		return -1, err
// 	}

// 	// adds a project to the mock filesystem
// 	projectPath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(projectId), project.Name)
// 	projectDataPath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(projectId), DATA_DIR_NAME, project.Name)
// 	if err = os.MkdirAll(projectPath, DIR_PERMISSIONS); err != nil {
// 		return -1, err
// 	}
// 	if err = os.MkdirAll(projectDataPath, DIR_PERMISSIONS); err != nil {
// 		return -1, err
// 	}

// 	project.Id = projectId
// 	return projectId, nil
// }

// /*
// adds a file to the db and filesystem given a file object, a project name, and a project id

// Params:
// 	file (*File) : a file struct to add to the db
// 	projectId (int) : the id of the project which this file is a part of
// 	projectName (string) : the 
// Returns:
// 	(int) : the file's id, assigned upon being added to the db (-1 if the operation is unsuccessful)
// 	(error) : an error if one occurs, nil otherwise
// */
// func addFile(file *File, data *CodeFileData, projectId int, projectName string) (int, error) {
// 	// inserts the file, getting the auto-generated file ID back from the query
// 	var fileId int
// 	var err error
// 	insertFile := fmt.Sprintf(INSERT_FILE,
// 		TABLE_FILES,
// 		getDbTag(&File{}, "ProjectId"),
// 		getDbTag(&File{}, "Path"))
// 	row := db.QueryRow(insertFile, projectId, file.Path)
// 	if row.Err() != nil {
// 		return -1, row.Err()
// 	}
// 	// gets the id from the just inserted file
// 	if err = row.Scan(&fileId); err != nil {
// 		return -1, err
// 	}

// 	// initializes the filesystem
// 	filePath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(projectId), projectName, file.Path)
// 	fileDataPath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(projectId), DATA_DIR_NAME, projectName, strings.TrimSuffix(file.Path, filepath.Ext(file.Path)) + ".json")

// 	// populates the filesystem with a test file and data about said test file
// 	testFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
// 	if err != nil {
// 		return -1, err
// 	}
// 	testDataFile, err := os.OpenFile(fileDataPath, os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
// 	if err != nil {
// 		return -1, err
// 	}

// 	// writes data to the file
// 	if _, err = testFile.Write([]byte(file.Content)); err != nil {
// 		return -1, err
// 	}
// 	jsonString, err := json.Marshal(data)
// 	if err != nil {
// 		return -1, err
// 	}
// 	if _, err = testDataFile.Write([]byte(jsonString)); err != nil {
// 		return -1, err
// 	}
// 	testFile.Close()
// 	testDataFile.Close()

// 	file.Id = fileId
// 	return fileId, nil
// }


// // function to test querying all projects at once
// func TestGetAllProjects(t *testing.T) {
// 	var err error

// 	// Set up server to listen with the getFile() function.
// 	muxRouter := mux.NewRouter()
// 	muxRouter.HandleFunc("/projects", getAllProjects) // TEMP: this path could change
// 	srv := &http.Server{Addr: ":"+TEST_SERVER_PORT, Handler: muxRouter}

// 	// Start server.
// 	go srv.ListenAndServe()

// 	/*
// 	test for basic functionality. Adds 2 projects to the db, then queries them and tests for equality
// 	*/
// 	func () {
// 		var projectId int // variable to temporarily store project ids as they are added to the db
// 		sentProjects := make(map[int]string) // variable to hold the id: project name mappings which are sent to the db

// 		// sets up the test environment (db and filesystem)		
// 		if err = initTestEnvironment(); err != nil {
// 			t.Errorf("Error initializing the test environment %s", err)
// 		}
// 		// uses a slice here so that we can add more projects to testProjects without breaking the test
// 		for _, proj := range testProjects[0:2] {
// 			projectId, err = addProject(proj)
// 			if err != nil {
// 				t.Errorf("Error adding project %s: %s", proj.Name, err)
// 			}
// 			// saves the added project with its id
// 			sentProjects[projectId] = proj.Name
// 		}

// 		// builds and sends and http get request
// 		resp, err := http.Get(fmt.Sprintf("%s:%s/projects", TEST_URL, TEST_SERVER_PORT))
// 		defer resp.Body.Close()
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		// checks the returned list of projects for equality with the sent list
// 		returnedProjects := make(map[int]string)
// 		json.NewDecoder(resp.Body).Decode(&returnedProjects)

// 		// tests that the proper values have been returned
// 		for k, v := range returnedProjects {
// 			if (v != sentProjects[k]) {
// 				t.Errorf("Projects of ids: %d do not have matching names. Given: %s, Returned: %s ", k, sentProjects[k], v)
// 			}
// 		}
		
// 		// destroys the test environment
// 		if err = clearTestEnvironment(); err != nil {
// 			t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
// 		}
// 	}()

// 	// Close server.
// 	if err = srv.Shutdown(context.Background()); err != nil {
// 		t.Errorf("HTTP server shutdown: %v", err)
// 	}
// }

// // function to test querying projects from the db and filesystem
// func TestGetProject(t *testing.T) {
// 	var err error

// 	// Set up server to listen with the getFile() function.
// 	muxRouter := mux.NewRouter()
// 	muxRouter.HandleFunc("/project", getProject) // TEMP: this path could change
// 	srv := &http.Server{Addr: ":"+TEST_SERVER_PORT, Handler: muxRouter}

// 	// Start server.
// 	go srv.ListenAndServe()

// 	/*
// 	Tests the basic ability of the CodeFiles module to load a project from the
// 	db and filesystem
// 	*/
// 	func() {
// 		var projectId int // holds the project id as returned from the addProject() function
// 		testFile := testFiles[0] // defines the file to use for the test here so that it can be easily changed
// 		testProject := testProjects[0] // defines the project to use for the test here so that it can be easily changed

// 		// initializes the filesystem and db
// 		if err = initTestEnvironment(); err != nil {
// 			t.Errorf("Error initializing the test environment %s", err)
// 		}
// 		// adds the test project to the filesystem and database
// 		projectId, err = addProject(testProject)
// 		if err != nil {
// 			t.Errorf("Error adding project %s: %s", testProject.Name, err)
// 		}
// 		// adds the test file to the filesystem and database
// 		_, err = addFile(testFile, testFileData[0], projectId, testProject.Name)
// 		if err != nil {
// 			t.Errorf("Error adding file %s: %s", testFile.Name, err)
// 		}
			
// 		// creates a request to get a project of a given id
// 		client := &http.Client{}
// 		req, err := http.NewRequest("GET", fmt.Sprintf("%s:%s/project", TEST_URL, TEST_SERVER_PORT), nil)
// 		if err != nil {
// 			t.Errorf("Error Retrieving Project: %v\n", err)
// 		}
// 		// sets a custom header of "project":id to query the specific project id
// 		req.Header.Set("project", fmt.Sprint(testProject.Id))
// 		resp, err := client.Do(req)
// 		defer resp.Body.Close()
// 		if err != nil {
// 			t.Errorf("Error while sending Get request: %v", err)
// 		}
// 		// if an error occurred while getting the file, it is printed out here
// 		if resp.StatusCode != http.StatusOK {
// 			t.Errorf("Error: %d", resp.StatusCode)
// 		}

// 		// marshals the json response into a Project struct
// 		project := &Project{}
// 		err = json.NewDecoder(resp.Body).Decode(&project)
// 		if err != nil {
// 			t.Error("Error while decoding server response: ", err)
// 		}

// 		// tests that the project matches the passed in data
// 		if (testProject.Id != project.Id) {
// 			t.Errorf("Project IDs do not match. Given: %d != Returned: %d", testProject.Id, project.Id)
// 		} else if (testProject.Name != project.Name) {
// 			t.Errorf("Project Names do not match. Given: %s != Returned: %s", testProject.Name, project.Name)
// 		// tests that the reviewers, authors, and file paths match (done directly here as there is only one constituent file)
// 		} else if (testProject.FilePaths[0] != project.FilePaths[0]) {
// 			t.Errorf("Project file path lists do not match. Given: %s != Returned: %s", testProject.FilePaths[0], project.FilePaths[0])
// 		}

// 		// destroys the filesystem and db
// 		if err = clearTestEnvironment(); err != nil {
// 			t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
// 		}
// 	}()

// 	// Close server.
// 	if err = srv.Shutdown(context.Background()); err != nil {
// 		fmt.Printf("HTTP server shutdown: %v", err)
// 	}
// }

// // function to test querying files
// func TestGetFile(t *testing.T) {
// 	var err error

// 	// Set up server to listen with the getFile() function.
// 	muxRouter := mux.NewRouter()
// 	muxRouter.HandleFunc("/project/file", getFile) // TEMP: this path could change
// 	srv := &http.Server{Addr: ":"+TEST_SERVER_PORT, Handler: muxRouter}

// 	// Start server.
// 	go srv.ListenAndServe()

// 	/*
// 	Tests the basic ability of the CodeFiles module to load the data from a
// 	valid file id passed to it. Simple valid one code file project
// 	*/
// 	func() {
// 		var projectId int // stores project id returned by addProject()
// 		var fileId int // stores the file id returned by addFile()
// 		testFile := testFiles[0] // the test file to be added to the db and filesystem (saved here so it can be easily changed)
// 		testProject := testProjects[0] // the test project to be added to the db and filesystem (saved here so it can be easily changed)

// 		// initializes the filesystem and db
// 		if err = initTestEnvironment(); err != nil {
// 			t.Errorf("Error initializing the test environment %s", err)
// 		}
// 		// adds a project to the database and filesystem
// 		projectId, err = addProject(testProject)
// 		if err != nil {
// 			t.Errorf("Error adding project %s: %v", testProject.Name, err)
// 		}
// 		// adds a file to the database and filesystem
// 		fileId, err = addFile(testFile, testFileData[0], projectId, testProject.Name)
// 		if err != nil {
// 			t.Errorf("Error adding file %s: %v", testFile.Name, err)
// 		}
// 		// sets the project id of the added file to link it with the project on this end
// 		testFile.ProjectId = projectId
			
// 		// creates a request to get a file of a given id
// 		client := &http.Client{}
// 		req, err := http.NewRequest("GET", fmt.Sprintf("%s:%s/project/file", TEST_URL, TEST_SERVER_PORT), nil)
// 		if err != nil {
// 			t.Errorf("Error creating request: %v\n", err)
// 		}
// 		// sets a custom header "file": file id to indicate which file is being queried to the server
// 		req.Header.Set("file", fmt.Sprint(fileId))
// 		resp, err := client.Do(req)
// 		defer resp.Body.Close()
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		// if an error occurred while querying, it's status code is printed here
// 		if resp.StatusCode != http.StatusOK {
// 			t.Errorf("Error: %d", resp.StatusCode)
// 		}

// 		// marshals the json response into a file struct
// 		file := &File{}
// 		err = json.NewDecoder(resp.Body).Decode(&file)
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		// tests that the file id of the returned struct is the same as that which was used for the query
// 		fileContent, err := base64.StdEncoding.DecodeString(file.Content)
// 		if err != nil {
// 			t.Error("unable to decode base64 file content")
// 		}

// 		// tests that the file id is correct
// 		if (testFile.Id != file.Id) {
// 			t.Errorf("File ID %d != %d", file.Id, testFile.Id)
// 		// tests for file name correctness
// 		} else if (testFile.ProjectId != file.ProjectId) {
// 			t.Errorf("File Project Id %d != %d", file.ProjectId, testFile.ProjectId)
// 		// tests if the file paths are identical
// 		} else if (testFile.Path != file.Path) {
// 			t.Errorf("File Path %s != %s", file.Path, testFile.Path)		
// 		// tests that the file content is correct
// 		} else if (testFile.Content != string(fileContent)) {
// 			t.Error("File Content does not match")
// 		}

// 		// destroys the filesystem and db
// 		if err = clearTestEnvironment(); err != nil {
// 			t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
// 		}
// 	}()

// 	// Close server.
// 	if err = srv.Shutdown(context.Background()); err != nil {
// 		fmt.Printf("HTTP server shutdown: %v", err)
// 	}
// }
