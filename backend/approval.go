// =======================================================================
// approval.go
// Authors: 190010425
// Created: February 28, 2022
//
// This file takes care of everything having to do w/ submission approval
// This includes review upload, editor approval, etc.
// =======================================================================

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	ENPOINT_REVIEW = "/review"
)

// ------------
// Router Functions
// ------------

// router function for reviewer review upload
func uploadReview(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] uploadReview request received from %v", r.RemoteAddr)
	resp := &StandardResponse{}
	reqBody := &UploadReviewBody{}

	// gets the submission ID from the vars
	params := mux.Vars(r)
	submissionID64, err := strconv.ParseUint(params["id"], 10, 32)
	submissionID := uint(submissionID64)
	if err != nil {
		resp = &StandardResponse{Message: "Given Submission ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// decodes request body
	} else if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
		resp = &StandardResponse{Message: "Unable to parse request body.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// validates reqBody format
	} else if err := validate.Struct(reqBody); err != nil {
		resp = &StandardResponse{Message: "Request body could not be validated.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// checks that the client is logged in
	} else if r.Context().Value("userId") == nil {
		// User is not validated - error out.
		resp = &StandardResponse{Message: "The client is unauthorized from making this query.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	// adds review and handles error cases
	} else {
		// builds review to add now that all fields have been validated
		review := &Review{
			ReviewerID: r.Context().Value("userId").(string),
			Approved: reqBody.Approved,
			Base64Value: reqBody.Base64Value,
		}
		// adds the review and formats response based upon error type if one occurs
		if err := addReview(review, submissionID); err != nil {
			switch err.(type) {
			case *NotReviewerError:
				log.Printf("[ERROR] %v\n", err.Error())
				resp = &StandardResponse{Message: "User is not a reviewer", Error: true}
				w.WriteHeader(http.StatusUnauthorized)

			case *DuplicateReviewError:
				log.Printf("[ERROR] %v\n", err.Error())
				resp = &StandardResponse{Message: "Reviewers can only upload one review each.", Error: true}
				w.WriteHeader(http.StatusBadRequest)

			default: // Unexpected error - error out as server error.
				log.Printf("[ERROR] could not upload review: %v\n", err)
				resp = &StandardResponse{Message: "Internal Server Error - could not upload review", Error: true}
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}

	// Return response body after function successful.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !resp.Error {
		log.Print("[INFO] uploadReview request successful\n")
	}
}


// ------------
// Helper Functions
// ------------

// adds a review to a given submission's metadata 
// 
// Params:
// 	review (*Review) : the review to be added
// 	submissionID (uint) : the id of the submission for the review to be added to
// Return:
// 	(error) : an error if one occurs, nil otherwise
func addReview(review *Review, submissionID uint) error {
	// gets a given submission, and adds the review to it
	submission, err := getSubmission(submissionID)
	if err != nil {
		return err
	}

	// checks that the reviewer is assigned to the given submission (implicitly checks usertype)
	isReviewer := false
	for _, reviewer := range submission.Reviewers {
		if reviewer.ID == review.ReviewerID {
			isReviewer = true
		}
	}
	if !isReviewer {
		return &NotReviewerError{UserID: review.ReviewerID, SubmissionID: submissionID}
	}

	// checks that the reviewer has not already uploaded a review (loop is ok here because the number of reviews is small)
	for _, currReview := range submission.MetaData.Reviews {
		if review.ReviewerID == currReview.ReviewerID {
			return &DuplicateReviewError{ UserID:review.ReviewerID, SubmissionID: submissionID }
		}
	}

	// adds the review to the given submission
	submission.MetaData.Reviews = append(submission.MetaData.Reviews, review)
	return addMetaData(submission)
}