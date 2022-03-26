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
	"gorm.io/gorm"
)

const (
	ENDPOINT_COMMENT = "/comment"
	ENDPOINT_EDIT    = "/edit"
	ENDPOINT_DELETE  = "/delete"
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
		ParentID: req.ParentID, Base64Value: req.Base64Value, StartLine: req.StartLine, EndLine: req.EndLine}); err != nil {
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

// edit existing comment
// POST /file/{id}/comment/edit
func PostEditUserComment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resp *StandardResponse
	req := &EditCommentPostBody{}

	// gets context struct and validates it
	if ctx, ok := r.Context().Value("data").(*RequestContext); !ok || validate.Struct(ctx) != nil {
		resp = &StandardResponse{Message: "Bad Request - No user logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if err := json.NewDecoder(r.Body).Decode(req); err != nil || validate.Struct(req) != nil {
		resp = &StandardResponse{Message: "Bad Request - Request format is invalid.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

		// edits the comment using the given controller method
	} else if err := ControllerEditComment(req, ctx.ID); err != nil {
		switch err.(type) {
		case *CommentNotFoundError:
			resp = &StandardResponse{Message: "Given comment does not exist.", Error: true}
			w.WriteHeader(http.StatusBadRequest)
		case *WrongPermissionsError:
			resp = &StandardResponse{Message: "Cannot edit comments you did not author.", Error: true}
			w.WriteHeader(http.StatusUnauthorized)
		default:
			log.Printf("[ERROR] Comment edit failed: %v", err)
			resp = &StandardResponse{Message: "Internal Server Error - Comment edit failed.", Error: true}
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		resp = &StandardResponse{Message: "Comment Edit Successful", Error: false}
	}

	// Encode response - set as error if empty
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] JSON repsonse formatting failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// takes in the body of an edit comment request and returns an error if one occurs
func ControllerEditComment(r *EditCommentPostBody, authorID string) error {
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// gets the comment from the db if it exists
		comment := &Comment{Model: gorm.Model{ID: r.ID}}
		if res := tx.Model(comment).Find(comment); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return &CommentNotFoundError{ID: r.ID}
		} else if comment.AuthorID != authorID {
			return &WrongPermissionsError{userID: authorID}
		}
		// updates the comment content if the comment exists
		if err := tx.Model(comment).Update("base64_value", r.Base64Value).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// edit existing comment
// POST /file/{id}/comment/delete
func PostDeleteUserComment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resp *StandardResponse
	req := &DeleteCommentPostBody{}

	// gets context struct and validates it
	if ctx, ok := r.Context().Value("data").(*RequestContext); !ok || validate.Struct(ctx) != nil {
		resp = &StandardResponse{Message: "Bad Request - No user logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if err := json.NewDecoder(r.Body).Decode(req); err != nil || validate.Struct(req) != nil {
		resp = &StandardResponse{Message: "Bad Request - Request format is invalid.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

		// deletes the comment using the given controller method
	} else if err := ControllerDeleteComment(req, ctx.ID); err != nil {
		switch err.(type) {
		case *CommentNotFoundError:
			resp = &StandardResponse{Message: "Given comment does not exist.", Error: true}
			w.WriteHeader(http.StatusBadRequest)
		case *WrongPermissionsError:
			resp = &StandardResponse{Message: "Cannot delete comments you did not author.", Error: true}
			w.WriteHeader(http.StatusUnauthorized)
		default:
			log.Printf("[ERROR] Comment delete failed: %v", err)
			resp = &StandardResponse{Message: "Internal Server Error - Comment delete failed.", Error: true}
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		resp = &StandardResponse{Message: "Comment Deleted Successfully", Error: false}
	}

	// Encode response - set as error if empty
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] JSON repsonse formatting failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// takes in the body of an edit comment request and returns an error if one occurs
func ControllerDeleteComment(r *DeleteCommentPostBody, authorID string) error {
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// gets the comment from the db if it exists
		comment := &Comment{Model: gorm.Model{ID: r.ID}}
		if res := tx.Model(comment).Find(comment); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return &CommentNotFoundError{ID: r.ID}
		} else if comment.AuthorID != authorID {
			return &WrongPermissionsError{userID: authorID}
		}
		// assigns all child comments to the parent of the current comment
		if err := tx.Model(&Comment{}).Where("parent_id = ?", comment.ID).
			Update("parent_id", comment.ParentID).Error; err != nil {
			return err
		}
		// deletes the comment
		if err := tx.Model(&Comment{}).Delete(comment).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
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
