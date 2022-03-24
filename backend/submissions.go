// =============================================================================
// submissions.go
// Authors: 190010425
// Created: November 18, 2021
//
// This file handles over-arching functionality of writing/reading submissions
// to/from the database
// =============================================================================

package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	SUBROUTE_SUBMISSION  = "/submission"
	SUBROUTE_SUBMISSIONS = "/submissions"

	ENDPOINT_QUERY_SUBMISSIONS   = "/query"
	ENDPOINT_UPLOAD_SUBMISSION   = "/create"
	ENDPOINT_DOWNLOAD_SUBMISSION = "/download"
	ENDPOINT_GET_TAGS            = "/tags"

	ORDER_NIL        = 0
	ORDER_ASCENDING  = 1
	ORDER_DESCENDING = 2
)

// Describe mux routing for submission-based endpoints.
func getSubmissionsSubRoutes(r *mux.Router) {
	submission := r.PathPrefix(SUBROUTE_SUBMISSION).Subrouter()
	submission.Use(jwtMiddleware)
	submissions := r.PathPrefix(SUBROUTE_SUBMISSIONS).Subrouter()
	submissions.Use(jwtMiddleware)

	// Submission routes:
	// + /submission/{id} - Get given submission.
	// + /submission/{id}/download - Downloads a submission as a zip archive
	// + /submission/{id}/assignreviewers - Assign reviewers to a given submission (in approval.go)
	// + /submission/{id}/review - upload a review for a submission (in approval.go)
	// + /submission/{id}/approve - change submission status to approve/dissaprove (in approval.go)
	// + /submission/{id}/export/{groupNumber} - export submission to another journal in the supergroup (in journal.go)
	submission.HandleFunc("/{id}", RouteGetSubmission).Methods(http.MethodGet)
	submission.HandleFunc("/{id}"+ENDPOINT_DOWNLOAD_SUBMISSION, GetDownloadSubmission).Methods(http.MethodGet)
	submission.HandleFunc("/{id}"+ENDPOINT_ASSIGN_REVIEWERS, PostAssignReviewers).Methods(http.MethodPost, http.MethodOptions)
	submission.HandleFunc("/{id}"+ENPOINT_REVIEW, PostUploadReview).Methods(http.MethodPost, http.MethodOptions)
	submission.HandleFunc("/{id}"+ENDPOINT_CHANGE_STATUS, PostUpdateSubmissionStatus).Methods(http.MethodPost, http.MethodOptions)
	submission.HandleFunc("/{id}"+ENDPOINT_EXPORT_SUBMISSION+"/{groupNumber}", PostExportSubmission).Methods(http.MethodPost, http.MethodOptions)

	// Submissions routes:
	// + /submissions/tags - gets all available tags currently stored in the database
	// + /submissions/query - queries a list of submissions based upon parameters
	// + /submissions/create - Create a submissions
	submissions.HandleFunc(ENDPOINT_GET_TAGS, GetAvailableTags).Methods(http.MethodGet)
	submissions.HandleFunc(ENDPOINT_QUERY_SUBMISSIONS, GetQuerySubmissions).Methods(http.MethodGet)
	submissions.HandleFunc(ENDPOINT_UPLOAD_SUBMISSION, PostUploadSubmissionByZip).Methods(http.MethodPost, http.MethodOptions)
}

// ------
// Router Functions
// ------

// function to send a list of all available tags to the frontend so that users know which tags allow filtering
// GET /submissions/tags
func GetAvailableTags(w http.ResponseWriter, r *http.Request) {
	stdResp := StandardResponse{}

	// queries the tags from the database
	categories := []Category{}
	if dbResp := gormDb.Model(&Category{}).Find(&categories); dbResp.RowsAffected == 0 {
		stdResp = StandardResponse{Message: "No tags added yet", Error: false}
		w.WriteHeader(http.StatusNoContent)
	} else if dbResp.Error != nil {
		log.Printf("[ERROR] could not query submissions: %v\n", dbResp.Error)
		stdResp = StandardResponse{Message: "Internal Server Error - could not query submissions", Error: true}
		w.WriteHeader(http.StatusInternalServerError)
	}
	// parses tags into an array of strings
	tags := []string{}
	for _, category := range categories {
		tags = append(tags, category.Tag)
	}
	// builds response
	resp := &GetAvailableTagsResponse{
		StandardResponse: stdResp,
		Tags:             tags,
	}
	// sends a response to the client
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// function to query a list of submissions with a set of query parameters to filter/order the list
// GET /submissions/query
func GetQuerySubmissions(w http.ResponseWriter, r *http.Request) {
	var err error
	var stdResp StandardResponse
	var resp *QuerySubmissionsResponse
	var submissions []Submission

	// gets the request context if there is a user logged in
	if ctx, ok := r.Context().Value("data").(*RequestContext); ok && validate.Struct(ctx) != nil {
		stdResp = StandardResponse{Message: "Bad Request Context", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	} else if submissions, err = ControllerQuerySubmissions(r.URL.Query(), ctx); err != nil {
		switch err.(type) {
		case *BadQueryParameterError:
			stdResp = StandardResponse{Message: fmt.Sprintf("Bad Request - %s", err.Error()), Error: true}
			w.WriteHeader(http.StatusBadRequest)
		case *ResultSetEmptyError:
			stdResp = StandardResponse{Message: "No submissions fit search queries", Error: false}
			w.WriteHeader(http.StatusOK)
		default:
			log.Printf("[ERROR] could not query submissions: %v\n", err)
			stdResp = StandardResponse{Message: "Internal Server Error - could not query submissions", Error: true}
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		stdResp = StandardResponse{Message: "", Error: false}
	}
	// builds the full response from the error message
	resp = &QuerySubmissionsResponse{
		StandardResponse: stdResp,
		Submissions:      submissions,
	}

	// sends a response to the client
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Controller to do the work of actually building a query to get submissions from
// the database and order them. This function uses helper functions to add filtering
// clauses to the final query
//
// Params:
// 	queryParams (url.Values) : a mapping of query parameters to their values
// 	userType (int) : the usertype of the currently logged in user (if there is one)
// Returns:
// 	([]Submission) : an array of submissions with ID and Name set ordered based upon the query
// 	(error) : an error if one occurs
func ControllerQuerySubmissions(queryParams url.Values, ctx *RequestContext) ([]Submission, error) {
	var submissions []Submission
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		tx = tx.Model(&Submission{})
		// includes submissions with a given tag
		if len(queryParams["tags"]) > 0 {
			tx = tx.Where("id IN (?)", gormDb.Table(
				"categories_submissions").Select("submission_id").Where("category_tag IN ?", queryParams["tags"]))
		}
		// filters submissions by author
		if len(queryParams["authors"]) > 0 {
			tx = tx.Where("id IN (?)", gormDb.Table(
				"authors_submission").Select("submission_id").Where("global_user_id IN ?", queryParams["authors"]))
		}
		// filters submissions by reviewer
		if len(queryParams["reviewers"]) > 0 {
			tx = tx.Where("id IN (?)", gormDb.Table(
				"reviewers_submission").Select("submission_id").Where("global_user_id IN ?", queryParams["reviewers"]))
		}
		// RegEx filtering for submission name
		if len(queryParams["name"]) > 0 {
			tx = filterByName(tx, regexp.QuoteMeta(queryParams["name"][0]))
		}
		// orders result set
		if len(queryParams["orderBy"]) > 0 {
			orderBy := queryParams["orderBy"][0]
			if orderBy != "newest" && orderBy != "oldest" && orderBy != "alphabetical" {
				return &BadQueryParameterError{ParamName: "orderBy", Value: queryParams["orderBy"]}
			}
			tx = orderSubmissionQuery(tx, orderBy)
		}
		// filters by usertype. If usertype is nil, only show approved submissions
		tx = filterByUserType(tx, ctx)

		// selects fields and gets submissions
		if res := tx.Select("id, name").Find(&submissions); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return &ResultSetEmptyError{}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return submissions, nil
}

// uses SQL REGEX to filter the submissions returned based on their names
func filterByName(tx *gorm.DB, submissionName string) *gorm.DB {
	params := map[string]interface{}{"full": submissionName}
	whereString := "submissions.name REGEXP @full"
	// only adds multiple regex conditions if the name given is multiple words
	if wordList := strings.Fields(submissionName); len(wordList) > 1 {
		for index, field := range wordList {
			whereString = whereString + " OR submissions.name REGEXP @" + fmt.Sprint(index)
			params[fmt.Sprint(index)] = field
		}
	}
	return tx.Where(whereString, params)
}

// adds a piece to an sql query to order the results
func orderSubmissionQuery(tx *gorm.DB, orderBy string) *gorm.DB {
	// order of submissions
	if orderBy == "newest" {
		tx = tx.Order("submissions.created_at ASC")
	} else if orderBy == "oldest" {
		tx = tx.Order("submissions.created_at DESC")
	} else if orderBy == "alphabetical" {
		tx = tx.Order("submissions.Name")
	}
	return tx
}

// adds usertype filters to a submission query
func filterByUserType(tx *gorm.DB, ctx *RequestContext) *gorm.DB {
	if ctx == nil {
		tx = tx.Where("submissions.approved = ?", true)
	} else {
		if ctx.UserType == USERTYPE_PUBLISHER {
			tx = tx.Where("approved = ? OR id IN (?)", true,
				gormDb.Table("authors_submission").Select("submission_id").Where("global_user_id = ?", ctx.ID))
		} else if ctx.UserType == USERTYPE_REVIEWER {
			tx = tx.Where("approved = ? OR id IN (?)", true,
				gormDb.Table("reviewers_submission").Select("submission_id").Where("global_user_id = ?", ctx.ID))
		} else if ctx.UserType == USERTYPE_REVIEWER_PUBLISHER {
			tx = tx.Where("submissions.approved = ? OR id IN (?) OR id IN (?)", true,
				gormDb.Table("reviewers_submission").Select("submission_id").Where("global_user_id = ?", ctx.ID),
				gormDb.Table("authors_submission").Select("submission_id").Where("global_user_id = ?", ctx.ID))
		}
	}
	// note editors can see all submissions and hence no filter is added for editors
	return tx
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
			Abstract: b.Abstract,
		},
		Runnable: b.Runnable,
	}
}

// Router function to upload new submissions by a Zip file with the file contents.
func PostUploadSubmissionByZip(w http.ResponseWriter, r *http.Request) {
	var resp UploadSubmissionResponse
	var reqBody UploadSubmissionByZipBody

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		resp.Message = "Could not decode body to correct format - " + err.Error()
		resp.Error = true
		w.WriteHeader(http.StatusBadRequest)
	} else if ctx, ok := r.Context().Value("data").(*RequestContext); !ok ||
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
		case *SubmissionNotRunnableError:
			resp.Message = err.Error()
			w.WriteHeader(http.StatusBadRequest)
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
		log.Printf("[WARN] Empty body given - returning.")
		return 0, errors.New("Submission is empty")
	}
	if err := validate.Struct(r); err != nil {
		return 0, err
	}
	files, err := getFileArrayFromZipBase64(r.ZipBase64Value)
	if err != nil {
		return 0, err
	}
	submission := adaptBodyToSubmission(&UploadSubmissionBody{
		Name: r.Name, License: r.License,
		Authors: r.Authors, Reviewers: r.Reviewers,
		Abstract: r.Abstract, Tags: r.Tags,
		Files: files, Runnable: r.Runnable,
	})
	submissionID, err := addSubmission(submission)
	if err != nil {
		return 0, err
	}
	err = storeZip(r.ZipBase64Value, submissionID)
	if err != nil {
		return 0, err
	} else {
		return submissionID, nil
	}
}

// Send submission data to the frontend for display. ID included for file
// and comment queries.
// GET /submission/{id}
func RouteGetSubmission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// gets the submission ID from the URL parameters
	params := mux.Vars(r)
	var encodable interface{}

	// Check path, execute controller, check errors.
	var err error
	var submission *Submission
	submissionID64, err := strconv.ParseUint(params["id"], 10, 32)
	if err != nil {
		encodable = StandardResponse{Message: "Given ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)
	} else if submission, err = getSubmission(uint(submissionID64)); err != nil {
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

	// submission not approved, can only be displayed for editors, authors or reviewers
	if submission != nil && (submission.Approved == nil || !*submission.Approved) {
		if ctx, ok := r.Context().Value("data").(*RequestContext); ok && validate.Struct(ctx) != nil {
			encodable = &StandardResponse{Message: "Error getting request context", Error: true}
			w.WriteHeader(http.StatusBadRequest)
		} else if ctx == nil {
			encodable = &StandardResponse{Message: "Non-user cannot view unapproved submission", Error: true}
			w.WriteHeader(http.StatusUnauthorized)
		} else if ctx.UserType != USERTYPE_EDITOR {
			// if the user is not an editor, check that they are either a reviewer or author for the given submission
			var allowed = false
			for _, author := range submission.Authors {
				if author.ID == ctx.ID {
					allowed = true
					break
				}
			}
			if !allowed {
				for _, reviewer := range submission.Reviewers {
					if reviewer.ID == ctx.ID {
						allowed = true
						break
					}
				}
			}
			if !allowed {
				encodable = &StandardResponse{Message: "Not authorized to access the given submission", Error: true}
				w.WriteHeader(http.StatusUnauthorized)
			}
		}
	}

	// writes JSON data for the submission to the HTTP connection
	if err := json.NewEncoder(w).Encode(encodable); err != nil {
		log.Printf("[ERROR] error formatting response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Compresses a given submission and returns it to the frontend to be downloaded
// GET /submission/{id}/download
func GetDownloadSubmission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/zip")
	var zipContent []byte

	// gets the submission ID and calls the controller
	params := mux.Vars(r)
	submissionID64, err := strconv.ParseUint(params["id"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else if zipContent, err = ControllerDownloadSubmission(uint(submissionID64)); err != nil {
		switch err.(type) {
		case *NoSubmissionError: // The given submission doesn't exist
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Printf("[ERROR] Internal server error on submission download - %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	w.Write(zipContent)
}

// Controller for the
//
// Params:
// 	submissionID (uint) : the unique ID of the submission being downloaded
// Returns:
// 	(string) : a string of the zip file's contents
// 	(error) : an error if one occurs
func ControllerDownloadSubmission(submissionID uint) ([]byte, error) {
	var submission Submission
	res := gormDb.Limit(1).Find(&submission, submissionID)
	if res.Error != nil {
		return nil, res.Error
	} else if res.RowsAffected == 0 {
		return nil, &NoSubmissionError{ID: submissionID}
	}

	// creates the zip archive if it doesn't exist, retrieves it otherwise
	zipPath := filepath.Join(getSubmissionDirectoryPath(submission), "project.zip")

	// read byte-array from the zip file, encode it to base64 and return
	zipContent, err := os.ReadFile(zipPath)
	if err != nil {
		return nil, err
	}
	retVal := make([]byte, base64.StdEncoding.EncodedLen(len(zipContent)))
	base64.StdEncoding.Encode(retVal, zipContent)
	return retVal, nil
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
	// checks for run file if the submission is marked runnable
	if submission.Runnable {
		seenRunFile := false
		for _, file := range submission.Files {
			if match, err := regexp.MatchString("^run\\.sh", file.Path); err != nil {
				return 0, err
			} else if match {
				seenRunFile = true
				break
			}
		}
		if !seenRunFile {
			return 0, &SubmissionNotRunnableError{}
		}
	}
	err := gormDb.Transaction(func(tx *gorm.DB) error {
		// Database operations
		categories := submission.Categories
		submission.Categories = []Category{}
		if err := createSubmissionToDb(tx, categories, submission); err != nil {
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
		}
		if err := addMetaData(submission); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = os.RemoveAll(getSubmissionDirectoryPath(*submission))
		return 0, err
	}
	return submission.ID, nil
}

// Add submission's clauses to a submission.
func createSubmissionToDb(tx *gorm.DB, categories []Category, submission *Submission) error {
	if err := tx.Omit(clause.Associations).Create(submission).Error; err != nil {
		return err
	}
	if err := addAuthors(tx, submission.Authors, submission.ID); err != nil {
		return err
	}
	if err := addReviewers(tx, submission.Reviewers, submission.ID); err != nil {
		return err
	}
	if err := tx.Model(&submission).Association("Categories").Append(categories); err != nil {
		return err
	}
	return nil
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
	model := &Submission{}
	model.ID = s.ID
	if err := tx.Model(model).Association("Files").Append(s.Files); err != nil {
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
