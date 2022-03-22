// ========================================
// comments.go
// Authors: 190010425
// Created: February 28, 2022
//
// This file takes care of user commenting
// ========================================

package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// -----------
// Router Functions
// -----------

// upload comment router function. Takes in a POST request and
// uses it to add a comment to the given file
//
// Response Codes:
// 	200 : comment was added succesfully
// 	401 : if the request does not have the proper security token
// 	400 : if the comment was not sent in the proper format
// 	500 : if something else goes wrong in the backend
// Response Body: {ID: <comment ID>}
func uploadUserComment(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] uploadUserComment request received from %v.", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Initialise message, response, and parameters.
	var message string
	var authorCtx RequestContext
	var encodable interface{}
	req := &NewCommentPostBody{}

	// Check function parameters in path and body.
	fileID64, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		message = "File ID given is not a number."
		w.WriteHeader(http.StatusBadRequest)
		goto RETURN

		// gets context struct and validates it
	} else if ctx, ok := r.Context().Value("data").(RequestContext); !ok || validate.Struct(ctx) != nil {
		authorCtx = ctx
		message = "No user logged in"
		w.WriteHeader(http.StatusUnauthorized)

	} else if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		message = "Request format is invalid."
		w.WriteHeader(http.StatusBadRequest)
		goto RETURN
	}

	// Get author ID from request, and add comment.
	if commentID, err := addComment(&Comment{
		AuthorID: authorCtx.ID, FileID: uint(fileID64),
		ParentID: req.ParentID, Base64Value: req.Base64Value, LineNumber: req.LineNumber,
	}); err != nil {
		message = "Comment creation failed."
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		encodable = NewCommentResponse{ID: commentID}
		goto RETURN
	}

RETURN:
	// Encode response - set as error if empty
	if encodable == nil {
		encodable = StandardResponse{Message: message, Error: true}
	} else if err := json.NewEncoder(w).Encode(encodable); err != nil {
		log.Printf("[ERROR] JSON repsonse formatting failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Printf("[INFO] uploadUserComment request from %v successful.", r.RemoteAddr)
	}
	return
}

// -----------
// Helper Functions
// -----------

// Add a comment to a file
//
// Params:
//	comment (*Comment) : The comment struct to add to the file
//	fileID (uint) : the id of the file to add a comment to
// Returns:
//	(uint) : the id of the added comment
//	(error) : an error if one occurs, nil otherwise
func addComment(comment *Comment) (uint, error) {
	// Check parameters.
	if comment == nil {
		return 0, errors.New("Comment cannot be nil")
	} else if comment.AuthorID == "" {
		return 0, errors.New("The author must exist.")
	}
	// adds the comment to the comments table with foreign key fileId and parentID
	file := &File{}
	file.ID = comment.FileID
	if err := gormDb.Model(file).Association("Comments").Append(comment); err != nil {
		return 0, err
	}
	return comment.ID, nil
}
