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
	submissions.HandleFunc("/{id}"+ENDPOINT_APPROVE, RouteAssignReviewers).Methods(http.MethodPost, http.MethodOptions)
	submissions.HandleFunc("/{id}"+ENPOINT_REVIEW, RouteUploadReview).Methods(http.MethodPost, http.MethodOptions)
	submissions.HandleFunc("/{id}"+ENDPOINT_APPROVE, RouteUpdateSubmissionStatus).Methods(http.MethodPost, http.MethodOptions)
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

		// gets context struct
	} else if ctx, ok := r.Context().Value("data").(RequestContext); !ok || validate.Struct(ctx) != nil {
		log.Printf("[ERROR] Could not validate request body")
		resp.Message = "Request body could not be validated."
		resp.Error = true
		w.WriteHeader(http.StatusUnauthorized)

	} else if ctx.UserType != USERTYPE_PUBLISHER && ctx.UserType != USERTYPE_REVIEWER_PUBLISHER {
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
	} else if !resp.Error {
		log.Print("[INFO] uploadSubmission request successful\n")
	}
}

// Controller for the upload submission POST request.
func ControllerUploadSubmission(body UploadSubmissionBody) (uint, error) {
	if err := validate.Struct(body); err != nil {
		// Deal with validation errors - print all failed fields.
		if errors, ok := err.(validator.ValidationErrors); ok {
			log.Printf("[WARN] Given submission is invalid! ")
			return 0, errors
		} else {
			// Validation process failed.
			log.Printf("[ERROR] Validation failure! ")
			return 0, err
		}
	}
	submission := adaptBodyToSubmission(&body)

	if submissionID, err := addSubmission(submission); err != nil {
		return 0, err
	} else {
		return submissionID, nil
	}
}

// Adapt an upload submission body to a submission structure.
func adaptBodyToSubmission(b *UploadSubmissionBody) *Submission {
	authors, reviewers, categories := []GlobalUser{}, []GlobalUser{}, []Category{}
	for _, author := range b.Authors {
		authors = append(authors, GlobalUser{ID: author})
	}
	for _, reviewer := range b.Reviewers {
		reviewers = append(reviewers, GlobalUser{ID: reviewer})
	}
	for _, category := range b.Tags {
		categories = append(categories, Category{Tag: category})
	}
	return &Submission{
		Name: b.Name, License: b.License,
		Files: b.Files, Categories: categories,
		Authors: authors, Reviewers: reviewers,
		MetaData: &SubmissionData{
			Abstract: "test abstract",
		},
	}
}

// Router function to upload new submissions by a Zip file with the file contents.
func PostUploadSubmissionByZip(w http.ResponseWriter, r *http.Request) {
	resp := UploadSubmissionResponse{}
	reqBody := UploadSubmissionByZipBody{}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		resp.Message = "Incorrect submission fields"
		resp.Error = true
		w.WriteHeader(http.StatusBadRequest)
	} else if ctx, ok := r.Context().Value("data").(RequestContext); !ok ||
		validate.Struct(ctx) != nil {
		resp.Message = "The client is unauthorized from making such request - not logged in."
		resp.Error = true
		w.WriteHeader(http.StatusUnauthorized)
	} else if ut := ctx.UserType; ut != USERTYPE_PUBLISHER && ut != USERTYPE_REVIEWER_PUBLISHER {
		resp.Message = "The client is unauthorized from making such request - not a publisher."
		resp.Error = true
		w.WriteHeader(http.StatusUnauthorized)
	} else if submissionID, err := ControllerUploadSubmissionByZip(&reqBody); err != nil {
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

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] Error formatting response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Controller for the UploadSubmissionByZip POST route.
func ControllerUploadSubmissionByZip(r *UploadSubmissionByZipBody) (uint, error) {
	if r == nil {
		return 0, errors.New("Submission is empty")
	}
	files, err := getFileArrayFromZipBase64(r.ZipBase64Value)
	if err != nil {
		return 0, err
	}

	submission := adaptBodyToSubmission(&UploadSubmissionBody{
		Name: r.Name, License: r.License,
		Authors: r.Authors, Reviewers: r.Reviewers,
		Abstract: r.Abstract, Tags: r.Tags,
		Files: files,
	})
	if submissionID, err := addSubmission(submission); err != nil {
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
	if submission == nil {
		return 0, errors.New("Submission is empty.")
	}
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// Database operations
		categories := submission.Categories
		submission.Categories = []Category{}
		if err := tx.Omit(clause.Associations).Create(submission).Error; err != nil {
			return err
		} else if err := addAuthors(tx, submission.Authors, submission.ID); err != nil {
			return err
		} else if err := addReviewers(tx, submission.Reviewers, submission.ID); err != nil {
			return err
		} else if err := tx.Model(&submission).Association("Categories").Append(categories); err != nil {
			return err
		}

		// creates the directories to hold the submission in the filesystem
		submissionPath := getSubmissionDirectoryPath(*submission)
		if err := os.MkdirAll(submissionPath, DIR_PERMISSIONS); err != nil {
			return err
		}

		// Add files and metadata to the system.
		if err := addFiles(tx, submission); err != nil {
			return err
		} else if err := addMetaData(submission); err != nil {
			return err
		}
		return nil
	}); err != nil {
		// An error has occured - db has rolled back, now filesystem is rolled back.
		if submission != nil {
			_ = os.RemoveAll(getSubmissionDirectoryPath(*submission))
		}
		return 0, err
	}
	return submission.ID, nil
}

// Add collection of files to a new submission.
func addFiles(tx *gorm.DB, s *Submission) error {
	// Check if files are valid
	filepaths := make(map[string]struct{})
	for _, file := range s.Files {
		if _, exists := filepaths[file.Path]; !exists {
			filepaths[file.Path] = struct{}{}
		} else {
			return &DuplicateFileError{Path: file.Path}
		}
	}

	// Add files to the database
	if err := tx.Model(s).Association("Files").Append(s.Files); err != nil {
		return err
	}

	// Add files to directory struct.
	submissionPath := getSubmissionDirectoryPath(*s)
	for _, fileRef := range s.Files {
		fPath := filepath.Join(submissionPath, fmt.Sprint(fileRef.ID))
		if f, err := os.Create(fPath); err != nil {
			return err
		} else {
			defer f.Close()
			f.Write([]byte(fileRef.Base64Value))
		}
	}
	return nil
}

// Add a submission's metadata to the filesystem.
func addMetaData(s *Submission) error {
	// opens a JSON file for the submission metadata, and writes a SubmissionData struct to it
	submissionPath := getSubmissionDirectoryPath(*s)
	if f, err := os.OpenFile(filepath.Join(submissionPath, "data.json"), os.O_CREATE|os.O_WRONLY, FILE_PERMISSIONS); err != nil {
		return err
	} else {
		defer f.Close()
		if err := f.Truncate(0); err != nil {
			return err
		} else if err := json.NewEncoder(f).Encode(s.MetaData); err != nil {
			return err
		}
	}
	return nil
}

// Add array of registered authors with appropriate permissions to a submission's authorized authors attributes.
// Errors on first unregistered author or author with invalid permissions.
func addAuthors(tx *gorm.DB, authors []GlobalUser, submissionID uint) error {
	// Check if there is at least 1 author.
	switch {
	case authors == nil, len(authors) == 0:
		return errors.New("There must be at least 1 author")
	}
	return appendUsers(tx, authors, []int{USERTYPE_PUBLISHER, USERTYPE_REVIEWER_PUBLISHER}, "Authors", submissionID)
}

// Add array of registered reviewers with appropriate permissions to a submission's authorized reviewers attribute.
// Errors on first unregistered reviewer or reviewer with invalid permissions.
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
		if res := tx.Preload(clause.Associations).Find(submission, submissionID); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return &NoSubmissionError{ID: submissionID}
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
	if err := gormDb.Select("Name, created_at, ID").First(&submission, submissionID).Error; err != nil {
		return nil, err
	}

	// reads the data file into a string
	dataPath := filepath.Join(getSubmissionDirectoryPath(*submission), "data.json")
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
