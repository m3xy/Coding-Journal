package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	ENDPOINT_EXPORT_SUBMISSION = "/export"     // on submissions sub-router
	ENDPOINT_IMPORT_SUBMISSION = "/submission" // on the journal sub-router
	ENDPOINT_USER              = "/user"       // on the journal sub-router
)

var journalURLs map[int]string = map[int]string{
	23: "https://cs3099user23.host.cs.st-andrews.ac.uk/api/v1/supergroup",
	5:  "cs3099user05.host.cs.st-andrews.ac.uk/api/v1/supergroup",
	13: "https://cs3099user13.host.cs.st-andrews.ac.uk/api/v1/supergroup",
	26: "https://cs3099user26.host.cs.st-andrews.ac.uk/api/v1/supergroup",
	2:  "https://cs3099user02.host.cs.st-andrews.ac.uk/api/v1/supergroup",
	20: "https://cs3099user20.host.cs.st-andrews.ac.uk/api/v1/supergroup",
}

// Set of all supergroup-appliant controllers and routes
// Authors: 190014935, 190010425

func getJournalSubroute(r *mux.Router) {
	journal := r.PathPrefix(SUBROUTE_JOURNAL).Subrouter()

	journal.Use(journalMiddleWare)
	journal.HandleFunc(ENDPOINT_LOGIN, logIn).Methods(http.MethodPost, http.MethodOptions)
	journal.HandleFunc(ENDPOINT_IMPORT_SUBMISSION, PostImportSubmission).Methods(http.MethodPost, http.MethodOptions)
	journal.HandleFunc(ENDPOINT_USER, GetUsers).Methods(http.MethodGet)
	journal.HandleFunc(ENDPOINT_USER+"/{id}", GetUser).Methods(http.MethodGet)
}

// ----------
// Router Functions
// ----------

// Validate if given security token works.
// Params:
// 	Header: securityToken
// Return:
//  200: Success - security token valid.
//  401: Failure - security token invalid.
func tokenValidation(w http.ResponseWriter, r *http.Request) {
}

/*
 Log in to website, check credentials correctness.
 Content type: application/json
 Success: 200, Credentials are correct
 Failure: 401, Unauthorized
 Returns: userId
*/
func logIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from log in request.
	user := JournalLoginPostBody{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("[ERROR] JSON decoder failed on log in.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get User ID from local credentials check.
	uuid, _, status := GetLocalUserID(user)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}

	// Marshal JSON and insert it into the response.
	if err := json.NewEncoder(w).Encode(JournalLogInResponse{ID: uuid}); err != nil {
		log.Printf("[ERROR] JSON Response Encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// gets all users from our Journal as a list
// GET /user
func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// gets an array of users using GORM smart select fields
	users := []GlobalUser{}
	if err := gormDb.Model(&GlobalUser{}).Preload("User").Find(&users).Error; err != nil {
		log.Printf("[ERROR] SQL Query Error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// parses users into global format
	globUsers := make([]SupergroupUser, len(users))
	for i, u := range users {
		globUsers[i] = SupergroupUser{
			ID: u.ID, FirstName: u.FirstName, LastName: u.LastName,
			Email: u.User.Email, PhoneNumber: u.User.PhoneNumber,
			Organization: u.User.Organization,
		}
	}
	// sends response
	if err := json.NewEncoder(w).Encode(globUsers); err != nil {
		log.Printf("[ERROR] JSON response encoding failed: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// gets user profile by ID
// GET /user/{id}
func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// queries user using URL parameters
	vars := mux.Vars(r)
	user := GlobalUser{}
	if res := gormDb.Model(&GlobalUser{}).Preload("User").
		Where("id = ?", vars["id"]).Limit(1).Find(&user); res.Error != nil {
		log.Printf("[ERROR] SQL query error: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if res.RowsAffected == 0 {
		log.Printf("[WARN] No user linked to %s", vars["id"])
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// adapts local user to supergroup-compliant format
	globUser := SupergroupUser{
		ID: user.ID, FirstName: user.FirstName, LastName: user.LastName,
		Email: user.User.Email, PhoneNumber: user.User.PhoneNumber,
		Organization: user.User.Organization,
	}

	// sends response
	if err := json.NewEncoder(w).Encode(globUser); err != nil {
		log.Printf("[ERROR] JSON response encoding failed: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// router function to export submissions
// POST /submission/{id}/export/{groupNumber}
func PostExportSubmission(w http.ResponseWriter, r *http.Request) {
	resp := &StandardResponse{}

	// gets submission ID and group number from URL
	params := mux.Vars(r)
	submissionID64, err1 := strconv.ParseUint(params["id"], 10, 32)
	submissionID := uint(submissionID64)
	groupNumber, err2 := strconv.Atoi(params["groupNumber"])
	if err1 != nil {
		resp = &StandardResponse{Message: "Given Submission ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

		// checks that group number is valid (note that our group numbers go in intervals of 3 starting at 2 i.e. 2, 5, 8, 11...)
	} else if _, ok := journalURLs[groupNumber]; !ok || err2 != nil {
		resp = &StandardResponse{Message: fmt.Sprintf("Given group number: %d invalid", groupNumber), Error: true}
		w.WriteHeader(http.StatusBadRequest)

		// gets context struct and validates it
	} else if ctx, ok := r.Context().Value("data").(*RequestContext); !ok || validate.Struct(ctx) != nil {
		resp = &StandardResponse{Message: "Request Context not set, user not logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

		// checks that the client has the proper permisssions (i.e. is an editor)
	} else if ctx.UserType != USERTYPE_EDITOR {
		resp = &StandardResponse{Message: "The client must have editor permissions to export submissions.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

		// gets supergroup compliant submission and exports it
	} else {
		// gets the supergroup compliant submission
		globalSubmission, err := localToGlobal(submissionID)
		if err != nil {
			switch err.(type) {
			case *NoSubmissionError:
				resp = &StandardResponse{Message: "Bad Request - Submission does not exist", Error: true}
				w.WriteHeader(http.StatusBadRequest)
			default:
				log.Printf("[ERROR] could not export submission: %v\n", err)
				resp = &StandardResponse{Message: "Internal Server Error - could not export submission", Error: true}
				w.WriteHeader(http.StatusInternalServerError)
			}
		}

		// makes request to export the submission
		var reqBody []byte
		var req *http.Request
		var globalResp *http.Response
		if reqBody, err = json.Marshal(globalSubmission); err != nil {
			goto INTERNAL_ERROR
		} else if req, err = http.NewRequest(http.MethodPost,
			journalURLs[groupNumber]+SUBROUTE_JOURNAL+"/submission", bytes.NewBuffer(reqBody)); err != nil {
			goto INTERNAL_ERROR
		} else if globalResp, err = sendSecureRequest(gormDb, req, groupNumber); err != nil {
			goto INTERNAL_ERROR
		} else {
			goto SUCCESS
		}
	INTERNAL_ERROR: // procedure common to any internal server error
		log.Printf("[ERROR] could not export submission: %v\n", err)
		resp = &StandardResponse{Message: "Internal Server Error - could not export submission", Error: true}
		w.WriteHeader(http.StatusInternalServerError)
	SUCCESS:
		w.WriteHeader(globalResp.StatusCode)
	}

	// Return response body after function successful.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// router function to import submissions from another Journal
// POST /journal/submission
func PostImportSubmission(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	reqBody := &SupergroupSubmission{}

	if r.Body == nil {
		resp = &StandardResponse{Message: "Request body empty.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

		// decodes request body and validates it
	} else if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil || validate.Struct(reqBody) != nil {
		resp = &StandardResponse{Message: "Unable to parse request body.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	} else {
		// converts reqBody (a supergroup-compliant submission) to a local format
		localSubmission, err := globalToLocal(reqBody)
		if err != nil {
			log.Printf("[ERROR] could not export submission: %v\n", err)
			resp = &StandardResponse{Message: "Internal Server Error - could not export submission", Error: true}
			w.WriteHeader(http.StatusInternalServerError)
		}

		// adds the local submission to the db
		if submissionID, err := addSubmission(localSubmission); err != nil {
			switch err.(type) {
			case *DuplicateFileError:
				resp = &StandardResponse{Message: err.Error(), Error: true}
				w.WriteHeader(http.StatusBadRequest)

			case *BadUserError, *WrongPermissionsError:
				resp = &StandardResponse{Message: err.Error(), Error: true}
				w.WriteHeader(http.StatusUnauthorized)

			default: // Unexpected error - error out as server error.
				log.Printf("[ERROR] could not import submission: %v\n", err)
				resp = &StandardResponse{Message: "Internal Server Error - could not import submission", Error: true}
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			resp = &UploadSubmissionResponse{
				StandardResponse: StandardResponse{Message: "Submission import successful!", Error: false},
				SubmissionID:     submissionID,
			}
		}
	}

	// Return response body after function successful.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ----------
// Helper Functions
// ----------

// This function queries a submission in the local format from the db, and transforms
// it into the supergroup compliant format
//
// Parameters:
// 	submissionID (int) : the id of the submission to be converted to the supergroup-compliant
// 		format
// Returns:
// 	(*SupergroupSubmission) : a supergroup compliant submission struct
// 	(error) : an error if one occurs
func localToGlobal(submissionID uint) (*SupergroupSubmission, error) {
	// gets the submission struct which submissionID refers to
	localSubmission, err := getSubmission(submissionID)
	if err != nil {
		return nil, err
	}

	// adds author names to an array
	authors := []SuperGroupAuthor{}
	for _, author := range localSubmission.Authors {
		authors = append(authors, SuperGroupAuthor{ID: author.ID, Journal: "11"})
	}
	// constructs an array of tags for the submission
	categories := []string{}
	for _, category := range localSubmission.Categories {
		categories = append(categories, category.Tag)
	}
	// creates the Supergroup metadata struct
	supergroupData := SupergroupSubmissionData{
		CreationDate: localSubmission.CreatedAt,
		Abstract:     localSubmission.MetaData.Abstract,
		License:      localSubmission.License,
		Categories:   categories,
		Authors:      authors,
	}

	// creates the list of file structs using the file paths and files.go
	var base64 string
	var supergroupFile SupergroupFile
	supergroupFiles := []SupergroupFile{}
	for _, file := range localSubmission.Files {
		fullFilePath := filepath.Join(getSubmissionDirectoryPath(*localSubmission), fmt.Sprint(file.ID))
		base64, err = getFileContent(fullFilePath)
		if err != nil {
			return nil, err
		}
		supergroupFile = SupergroupFile{
			Name:        file.Path,
			Base64Value: base64,
		}
		supergroupFiles = append(supergroupFiles, supergroupFile)
	}

	// creates the supergroup submission to return
	return &SupergroupSubmission{
		Name:     localSubmission.Name,
		MetaData: supergroupData,
		CodeVersions: []SupergroupCodeVersion{
			{TimeStamp: localSubmission.CreatedAt,
				Files: supergroupFiles},
		},
	}, nil
}

// This function takes in a supergroup compliant submission, and uses it to construct a valid
// local-format submission
//
// Parameters:
// 	globalSubmission (SupergroupSubmission) : The supergroup compliant submission to be converted
// Returns:
// 	(*Submission) : a local submission struct
// 	(error) : an error if one occurs
func globalToLocal(globalSubmission *SupergroupSubmission) (*Submission, error) {
	if globalSubmission == nil {
		return nil, errors.New("global submission cannot be nil")
	}
	// builds the array of local file objects from the global submission's files
	files := []File{}
	for _, file := range globalSubmission.CodeVersions[0].Files {
		files = append(files, File{
			Path:        file.Name,
			Base64Value: file.Base64Value,
		})
	}
	// builds the array of authors
	authors := []GlobalUser{}
	for _, author := range globalSubmission.MetaData.Authors {
		authors = append(authors, GlobalUser{ID: author.ID})
	}
	// builds the array of tags
	categories := []Category{}
	for _, category := range globalSubmission.MetaData.Categories {
		categories = append(categories, Category{Tag: category})
	}

	// constructs and returns the final submission
	return &Submission{
		Name:       globalSubmission.Name,
		License:    globalSubmission.MetaData.License,
		Files:      files,
		Authors:    authors,
		Reviewers:  []GlobalUser{},
		Categories: categories,
		MetaData:   &SubmissionData{Abstract: globalSubmission.MetaData.Abstract},
	}, nil
}
