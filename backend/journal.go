package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

// Set of all supergroup-appliant controllers and routes
// Authors: 190014935

func getJournalSubroute(r *mux.Router) {
	journal := r.PathPrefix(SUBROUTE_JOURNAL).Subrouter()

	journal.Use(journalMiddleWare)
	journal.HandleFunc(ENDPOINT_LOGIN, logIn).Methods(http.MethodPost, http.MethodOptions)
}

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
	categories := []string{}
	fmt.Printf("%v\n", localSubmission.Categories)
	for _, category := range localSubmission.Categories {
		categories = append(categories, category.Tag)
	}
	// creates the Supergroup metadata struct
	supergroupData := &SupergroupSubmissionData{
		CreationDate: fmt.Sprint(localSubmission.CreatedAt),
		Categories:   categories,
		Abstract:     localSubmission.MetaData.Abstract,
		License:      localSubmission.License,
	}

	// adds author names to an array
	authorNames := []string{}
	for _, author := range localSubmission.Authors {
		authorNames = append(authorNames, author.FullName)
	}
	supergroupData.AuthorNames = authorNames

	// creates the list of file structs using the file paths and files.go
	var base64 string
	var supergroupFile *SupergroupFile
	supergroupFiles := []*SupergroupFile{}
	for _, file := range localSubmission.Files {
		fullFilePath := filepath.Join(getSubmissionDirectoryPath(*localSubmission), fmt.Sprint(file.ID))
		base64, err = getFileContent(fullFilePath)
		if err != nil {
			return nil, err
		}
		supergroupFile = &SupergroupFile{
			Name:        filepath.Base(file.Path),
			Base64Value: base64,
		}
		supergroupFiles = append(supergroupFiles, supergroupFile)
	}

	// creates the supergroup submission to return
	return &SupergroupSubmission{
		Name:     localSubmission.Name,
		Files:    supergroupFiles,
		MetaData: supergroupData,
	}, nil
}
