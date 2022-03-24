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

// upload comment router function.
// POST /file/{id}/newcomment
func PostUploadUserComment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := &NewCommentResponse{}
	req := &NewCommentPostBody{}

	// Check function parameters in path and body.
	fileID64, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		resp.StandardResponse = StandardResponse{Message: "Bad Request - Given File ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// gets context struct and validates it
	} else if ctx, ok := r.Context().Value("data").(*RequestContext); !ok || validate.Struct(ctx) != nil {
		resp.StandardResponse = StandardResponse{Message: "Bad Request - No user logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if err := json.NewDecoder(r.Body).Decode(req); err != nil || validate.Struct(req) != nil {
		resp.StandardResponse = StandardResponse{Message: "Bad Request - Request format is invalid.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// creates the comment using the given helper method
	} else if commentID, err := addComment(&Comment{AuthorID: ctx.ID, FileID: uint(fileID64),
		ParentID: req.ParentID, Base64Value: req.Base64Value, LineNumber: req.LineNumber}); err != nil {
		log.Printf("[ERROR] Comment creation failed: %v", err)
		resp.StandardResponse = StandardResponse{Message: "Internal Server Error - Comment creation failed.", Error: true}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.ID = commentID
	}

	// Encode response - set as error if empty
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] JSON repsonse formatting failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
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
	// Check parameters
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
