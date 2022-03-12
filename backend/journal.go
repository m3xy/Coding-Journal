package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"bytes"

	"github.com/gorilla/mux"
)

const (
	ENDPOINT_EXPORT_SUBMISSION = "/export" // on submissions sub-router
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
	log.Printf("[INFO] Token validation from %v successful.", r.RemoteAddr)
}

/*
 Log in to website, check credentials correctness.
 Content type: application/json
 Success: 200, Credentials are correct
 Failure: 401, Unauthorized
 Returns: userId
*/
func logIn(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received log in request from %v", r.RemoteAddr)
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
	log.Printf("[INFO] log in from %s at email %s successful.", r.RemoteAddr, user.Email)
}


// router function to export submissions
// POST /submission/{id}/export/{groupNumber}
func PostExportSubmission(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] ExportSubmission request received from %v", r.RemoteAddr)
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
	} else if ctx, ok := r.Context().Value("data").(RequestContext); !ok || validate.Struct(ctx) != nil {
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
	} else if !resp.Error {
		log.Print("[INFO] AssignReviewers request successful\n")
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
		authors = append(authors, SuperGroupAuthor{ID: author.ID, Journal:"11"})
	}
	// constructs an array of tags for the submission
	categories := []string{}
	for _, category := range localSubmission.Categories {
		categories = append(categories, category.Tag)
	}
	// creates the Supergroup metadata struct
	supergroupData := &SupergroupSubmissionData{
		CreationDate: fmt.Sprint(localSubmission.CreatedAt),
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
