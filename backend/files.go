// // ================================================================================
// // files.go
// // Authors: 190010425
// // Created: November 23, 2021
// //
// // This file handles reading/writing code files along with their
// // data.
// //
// // The directory structure for the filesystem is as follows:
// //
// // Submission ID (as stored in db submissions table)
// // 	> <submission_name>/ (as stored in the submissions table)
// // 		... (submission directory structure)
// // 	> .data/
// // 		> submission_data.json
// // 		... (submission directory structure)
// // notice that in the filesystem, the .data dir structure mirrors the
// // submission, so that each file in the submission can have a .json file storing
// // its data which is named in the same way as the source code (the only difference
// // being the extension)
// // ================================================================================

package main

// import (
	// "encoding/json"
	// "errors"
	// "fmt"
	// "io/ioutil"
	// "log"
	// "net/http"
	// "os"
	// "path/filepath"
	// "strconv"
	// "strings"
	// "time"
// )

// file constants, includes
const (
	// TEMP: hard coded for testing
	FILESYSTEM_ROOT = "../filesystem/" // path to the root directory holding all submission directories TEMP: maybe set with an env variable?
	DATA_DIR_NAME   = ".data"          // name of the hidden data dir to be put into the submission directory structure

	// File Mode Constants
	DIR_PERMISSIONS  = 0755 // permissions for filesystem directories
	FILE_PERMISSIONS = 0644 // permissions for submission files
)

// // -----
// // Router functions
// // -----

// // Upload lone code file to system. File is wrapped to dummy submission with same name.
// //
// // TODO: Replace this function with a generalized submission upload method in the submissions.go file
// //
// // Responses:
// //	- 200 : if action completes successfully
// // 	- 401 : if request does not have the proper security token
// // 	- 400 : for malformatted request
// // 	- 500 : if something goes wrong in the backend
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
// func uploadSingleFile(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("uploadSingleFile request received from %v", r.RemoteAddr)

// 	w.Header().Set("Content-Type", "application/json")
// 	var request map[string]interface{}
// 	json.NewDecoder(r.Body).Decode(&request)

// 	// Parse data into local variables
// 	fileName := request[getJsonTag(&File{}, "Name")]			// file name as a string
// 	fileAuthor := request["author"]                        		// author's user Id
// 	fileContent := request[getJsonTag(&File{}, "Base64Value")]	// base64 encoding of file content

// 	// Put parsed values into a file object and a submission object
// 	wrapperSubmission := &Submission{
// 		Name:      fileName.(string),
// 		Authors:   []string{fileAuthor.(string)},
// 		Reviewers: []string{},
// 		MetaData: &SubmissionData{
// 			Abstract: "",
// 			Reviews:  comments,
// 		},
// 	}
// 	file := &File{
// 		SubmissionName: fileName.(string),
// 		Path:           fileName.(string),
// 		Name:           fileName.(string),
// 		Base64Value:	fileContent.(string),
// 	}

// 	// adds file to the db and filesystem
// 	submissionId, err := addSubmission(wrapperSubmission)
// 	if err != nil {
// 		log.Printf("[ERROR] Submission creation failure: %v", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	_, err = addFileTo(file, submissionId)
// 	if err != nil {
// 		log.Printf("[ERROR] File import failure: %v", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	wrapperSubmission.Id = submissionId

// 	// writes the wraper submission as a response
// 	jsonString, err := json.Marshal(wrapperSubmission)
// 	if err != nil {
// 		log.Printf("[ERROR] JSON formatting failure: %v", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	log.Printf("uploadSingleFile request from %v successful.", r.RemoteAddr)
// 	w.Write([]byte(jsonString))
// }

// // Retrieve code files from filesystem. Returns
// // file with comments and metadata. Receives
// // a FilePath and submissionId as header strings in
// // the request
// //
// // TEMP: this might go away due to the supergroup spec
// //
// // Response Codes:
// //	200 : File exists, getter success.
// //	401 : if the request does not have the proper security token
// //	400 : malformatted request, or non-existent submission/file id
// //	500 : if something else goes wrong in the backend
// // Response Body:
// // 		file: object
// // 			fileName: string
// //			filePath: string
// //			submissionName: string
// //			submissionId: int
// // 			base64Value: string
// // 			comments: array of objects
// // 				author: int
// //				time: datetime string
// //				base64Value: string
// //				replies: object (same as comments)
// func getFile(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("[INFO] getFile request received from %v", r.RemoteAddr)
// 	w.Header().Set("Content-Type", "application/json")

// 	// gets the submission Id from the URL parameters
// 	params := r.URL.Query()
// 	submissionId, err := strconv.Atoi(params[getJsonTag(&File{}, "SubmissionId")][0])
// 	if err != nil {
// 		log.Printf("[ERROR] Invalid submission ID: %s", params[getJsonTag(&File{}, "SubmissionId")][0])
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	filePath := params[getJsonTag(&File{}, "Path")][0]

// 	// queries the submission name from the database
// 	var submissionName string
// 	querySubmissionName := fmt.Sprintf(
// 		SELECT_ROW,
// 		getDbTag(&Submission{}, "Name"),
// 		TABLE_SUBMISSIONS,
// 		getDbTag(&Submission{}, "Id"),
// 	)
// 	row := db.QueryRow(querySubmissionName, submissionId)
// 	if err = row.Scan(&submissionName); err != nil {
// 		log.Printf("[ERROR] Database query failed: %v", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	// builds path to the file and it's corresponding data file using the queried submission name
// 	fullFilePath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), submissionName, filePath)
// 	fullDataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), DATA_DIR_NAME,
// 		submissionName, strings.Replace(filePath, filepath.Ext(filePath), ".json", 1))

// 	// constructs a file object to return to the frontend
// 	file := &File{
// 		SubmissionId:   submissionId,
// 		SubmissionName: submissionName,
// 		Path:           filePath,
// 		Name:           filepath.Base(filePath),
// 	}
// 	// gets file content and comments
// 	file.Base64Value, err = getFileContent(fullFilePath)
// 	if err != nil {
// 		log.Printf("[ERROR] Failed to retrieve file content: %v", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	file.MetaData, err = getFileMetaData(fullDataPath)
// 	if err != nil {
// 		log.Printf("[ERROR] Failed to retrieve file metadata: %v", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	// writes JSON data for the file to the HTTP connection if no error has occured
// 	response, err := json.Marshal(file)
// 	if err != nil {
// 		log.Printf("[ERROR] JSON formatting failed: %v", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	log.Printf("[INFO] getFile request from %v successful.", r.RemoteAddr)
// 	w.Write(response)
// }

// // upload comment router function. Takes in a POST request and
// // uses it to add a comment to the given file
// //
// // Response Codes:
// // 	200 : comment was added succesfully
// // 	401 : if the request does not have the proper security token
// // 	400 : if the comment was not sent in the proper format
// // 	500 : if something else goes wrong in the backend
// // Response Body: empty
// func uploadUserComment(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("[INFO] uploadUserComment request received from %v.", r.RemoteAddr)
// 	w.Header().Set("Content-Type", "application/json")

// 	// parses the json request body into a map
// 	var request map[string]interface{}
// 	err := json.NewDecoder(r.Body).Decode(&request)
// 	if err != nil {
// 		log.Printf("[ERROR] JSON decoding failed: %v", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	// gets submission Id and file path
// 	filePath := request[getJsonTag(&File{}, "Path")].(string)
// 	submissionId := int(request[getJsonTag(&File{}, "SubmissionId")].(float64))
// 	// Parse data into Comment structure
// 	comment := &Comment{
// 		AuthorId: request[getJsonTag(&Comment{}, "AuthorId")].(string), // authors user id
// 		Time:     fmt.Sprint(time.Now()),
// 		Base64Value:  request[getJsonTag(&Comment{}, "Base64Value")].(string),
// 		Replies:  nil, // replies are nil upon insertion
// 	}

// 	// gets the fileId from the database
// 	var fileId int
// 	queryFileId := fmt.Sprintf(
// 		SELECT_ROW_TWO_CONDITION,
// 		getDbTag(&File{}, "Id"),
// 		TABLE_FILES,
// 		getDbTag(&File{}, "SubmissionId"),
// 		getDbTag(&File{}, "Path"),
// 	)
// 	row := db.QueryRow(queryFileId, submissionId, filePath)
// 	if err = row.Scan(&fileId); err != nil {
// 		log.Printf("[ERROR] Database query failed: %v\n", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	// adds the comment to the file, returns code OK if successful
// 	if err = addComment(comment, fileId); err != nil {
// 		log.Printf("[ERROR] Comment creation failed: %v\n", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	log.Printf("uploadUserComment request from %v successful.", r.RemoteAddr)
// 	w.WriteHeader(http.StatusOK)
// }

// // -----
// // Helper Functions
// // -----

// // Add file into submission, and store it to FS and DB.
// // Note: Need valid submission. No comments exist on file
// // creation.
// //
// // Params:
// //	file (*File) : the file to add to the db and filesystem (all fields but Id and SubmissionId MUST be set)
// //	submissionId (int) : the id of the submission which the added file is to be linked
// //		to as an unsigned integer
// // Returns:
// //	(int) : the id of the added file (0 if an error occurs)
// //	(error) : if the operation fails
// func addFileTo(file *File, submissionId int) (int, error) {
// 	// declares return value variables
// 	var fileId int
// 	var err error

// 	// formats SQL query to insert the file into the db
// 	insertFile := fmt.Sprintf(
// 		INSERT_FILE,
// 		TABLE_FILES,
// 		getDbTag(&File{}, "SubmissionId"),
// 		getDbTag(&File{}, "Path"),
// 	)
// 	// executes the formatted query, returning the fileId
// 	// (note that here SQL implicitly checks that the submissionId exists in the submissions table via Foreign key constraint)
// 	row := db.QueryRow(insertFile, submissionId, file.Path)
// 	// gets the id from the just inserted file
// 	if err = row.Scan(&fileId); err != nil {
// 		return 0, err
// 	}

// 	// Add file to filesystem
// 	filePath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), file.SubmissionName, file.Path)
// 	fileDataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), DATA_DIR_NAME,
// 		file.SubmissionName, strings.Replace(file.Path, filepath.Ext(file.Path), ".json", 1))

// 	// file paths without the file name (to create dirs if they don't exist yet)
// 	fileDirPath := filepath.Dir(filePath)
// 	fileDataDirPath := filepath.Dir(fileDataPath)

// 	// mkdir files's dir in case they don't yet exist
// 	if err = os.MkdirAll(fileDirPath, DIR_PERMISSIONS); err != nil {
// 		return 0, err
// 	} else if err = os.MkdirAll(fileDataDirPath, DIR_PERMISSIONS); err != nil {
// 		return 0, err
// 	}

// 	// Create and open file and it's corresponding data file
// 	codeFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
// 	if err != nil {
// 		return 0, err
// 	}
// 	dataFile, err := os.OpenFile(
// 		fileDataPath, os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS)
// 	if err != nil {
// 		return 0, err
// 	}

// 	// writes the file content
// 	if _, err = codeFile.Write([]byte(file.Base64Value)); err != nil {
// 		return 0, err
// 	}
// 	// Writes given file's metadata
// 	jsonString, err := json.Marshal(file.MetaData)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if _, err = dataFile.Write([]byte(jsonString)); err != nil {
// 		return 0, err
// 	}

// 	// closes files
// 	codeFile.Close()
// 	dataFile.Close()

// 	// Operation was successful ==> file Id set in file object and file returned.
// 	file.Id = fileId
// 	file.SubmissionId = submissionId
// 	return fileId, nil
// }

// // Add a comment to a given file
// //
// // Params:
// //	comment (*Comment) : The comment struct to add to the file
// //	fileId (int) : the id of the file to add a comment to
// // Returns:
// //	(error) : an error if one occurs, nil otherwise
// func addComment(comment *Comment, fileId int) error {
// 	var err error
// 	// error cases
// 	if comment == nil {
// 		return errors.New("Comment cannot be nil")
// 	} else if fileId < 0 {
// 		return errors.New("File Id must be > 0")
// 	}

// 	// checks that the author of the comment is a registered user (either here or in another journal)
// 	var authorExists bool
// 	queryAuthorExists := fmt.Sprintf(
// 		SELECT_EXISTS,
// 		"*",
// 		TABLE_IDMAPPINGS,
// 		getDbTag(&IdMappings{}, "GlobalId"),
// 	)
// 	// executes the query, and returns an error if the author is not registered
// 	row := db.QueryRow(queryAuthorExists, comment.AuthorId)
// 	if err = row.Scan(&authorExists); err != nil {
// 		return err
// 	}
// 	if !authorExists {
// 		return errors.New("Authors of comments must be registered in the db")
// 	}

// 	// queries the database to get the file path so that the file's data file can be found
// 	var submissionId string
// 	var submissionName string
// 	var filePath string
// 	// builds a query to get the file's name, submission id, and it's submission's name
// 	columns := fmt.Sprintf(
// 		"%s, %s, %s",
// 		TABLE_SUBMISSIONS+"."+getDbTag(&Submission{}, "Id"),
// 		getDbTag(&Submission{}, "Name"),
// 		getDbTag(&File{}, "Path"),
// 	)
// 	queryPath := fmt.Sprintf(
// 		SELECT_ROW_INNER_JOIN,
// 		columns,
// 		TABLE_FILES,
// 		TABLE_SUBMISSIONS,
// 		TABLE_FILES+"."+getDbTag(&File{}, "SubmissionId"),
// 		TABLE_SUBMISSIONS+"."+getDbTag(&Submission{}, "Id"),
// 		TABLE_FILES+"."+getDbTag(&File{}, "Id"),
// 	)
// 	// executes the query and builds the file path if it was successful
// 	row = db.QueryRow(queryPath, fileId)
// 	if err = row.Scan(&submissionId, &submissionName, &filePath); err != nil {
// 		return err
// 	}
// 	dataFilePath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), DATA_DIR_NAME,
// 		submissionName, strings.Replace(filePath, filepath.Ext(filePath), ".json", 1))

// 	// reads the data file content and formats it into a FileData struct
// 	data := &FileData{}
// 	dataFileContent, err := ioutil.ReadFile(dataFilePath)
// 	if err = json.Unmarshal(dataFileContent, data); err != nil {
// 		return err
// 	}

// 	// adds the new comment and writes to the
// 	data.Comments = append(data.Comments, comment)
// 	dataBytes, err := json.Marshal(data)
// 	if err != nil {
// 		return err
// 	}

// 	// if everything has gone correctly, the new data is written to the file
// 	return ioutil.WriteFile(dataFilePath, dataBytes, FILE_PERMISSIONS)
// }

// // helper function to return a file object given its ID
// //
// // TODO : write tests for this function
// //
// // Params:
// // 	fileId (int) : the file's unique id
// // Returns:
// //	(*File) : the a file struct corresponding to the given ID
// // 	(error) : an error if something goes wrong
// func getFileData(fileId int) (*File, error) {
// 	// builds a query to get the file's name, submission id, and it's submission's name
// 	var submissionId int
// 	var submissionName string
// 	var filePath string
// 	columns := fmt.Sprintf(
// 		"%s, %s, %s",
// 		TABLE_SUBMISSIONS+"."+getDbTag(&Submission{}, "Id"),
// 		getDbTag(&Submission{}, "Name"),
// 		getDbTag(&File{}, "Path"),
// 	)
// 	queryPath := fmt.Sprintf(
// 		SELECT_ROW_INNER_JOIN,
// 		columns,
// 		TABLE_FILES,
// 		TABLE_SUBMISSIONS,
// 		TABLE_FILES+"."+getDbTag(&File{}, "SubmissionId"),
// 		TABLE_SUBMISSIONS+"."+getDbTag(&Submission{}, "Id"),
// 		TABLE_FILES+"."+getDbTag(&File{}, "Id"),
// 	)
// 	// executes the query and builds the file path if it was successful
// 	row := db.QueryRow(queryPath, fileId)
// 	if err := row.Scan(&submissionId, &submissionName, &filePath); err != nil {
// 		return nil, err
// 	}

// 	// builds path to the file and it's corresponding data file using the queried submission name
// 	fullFilePath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), submissionName, filePath)
// 	fullDataPath := filepath.Join(FILESYSTEM_ROOT, fmt.Sprint(submissionId), DATA_DIR_NAME,
// 		submissionName, strings.Replace(filePath, filepath.Ext(filePath), ".json", 1))

// 	// constructs a file object to return to the frontend
// 	file := &File{
// 		SubmissionId:   submissionId,
// 		SubmissionName: submissionName,
// 		Path:           filePath,
// 		Name:           filepath.Base(filePath),
// 	}
// 	// gets file content and metadata
// 	var err error
// 	file.Base64Value, err = getFileContent(fullFilePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	file.MetaData, err = getFileMetaData(fullDataPath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return file, nil
// }

// // // Gets a supergroup compliant file struct given the file id
// // //
// // // Params:
// // // 	fileId (int) : the unique of the given file
// // // Returns:
// // // 	(*SupergroupFile) : a supergroupp compliant file object
// // // 	(error) : an error if one occurs
// // func getSuperGroupFile(fileId int) (*SupergroupFile, error) {
// // 	// builds a query to get the file's name from the database
// // 	var fileName string
// // 	queryName := fmt.Sprintf(
// // 		SELECT_ROW,
// // 		getDbTag(&File{}, "Name"),
// // 		TABLE_FILES,
// // 		getDbTag(&File{}, "Id")
// // 	)
// // 	// executes the query and builds the file path if it was successful
// // 	row := db.QueryRow(queryPath, fileId)
// // 	if err := row.Scan(&fileName); err != nil {
// // 		return nil, err
// // 	}

// // 	// constructs a file object to return to the frontend
// // 	file := &SupergroupFile{
// // 		Name:           filepath.Base(filePath),
// // 	}
// // 	// gets file content and metadata
// // 	var err error
// // 	file.Base64Value, err = getFileContent(fullFilePath)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	file.MetaData, err = getFileMetaData(fullDataPath)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	return file, nil
// // }

// // Get file content from filesystem.
// // Params:
// // 	filePath (string): an absolute path to the file
// // Returns:
// // 	(error) : if something goes wrong, nil otherwise
// func getFileContent(filePath string) (string, error) {
// 	// reads in the file's content
// 	fileData, err := ioutil.ReadFile(filePath)
// 	if err != nil {
// 		return "", err
// 	}
// 	// if no error occurred, assigns file.Base64Value a value
// 	return string(fileData), nil
// }

// // Get a file's metadata from filesystem, and returns a FileData
// // objects
// //
// // Params:
// //	dataPath (string) : a path to the data file containing a given file's meta-data
// //Returns:
// //  (*FileData) : the file's metadata
// //	(error) : if something goes wrong, nil otherwise
// func getFileMetaData(dataPath string) (*FileData, error) {
// 	// reads the file contents into a json string
// 	jsonData, err := ioutil.ReadFile(dataPath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// fileData is parsed from json into the FileData struct
// 	fileData := &FileData{}
// 	err = json.Unmarshal(jsonData, fileData)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// if no error occurred, return metadata
// 	return fileData, nil
// }
