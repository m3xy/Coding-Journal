package main

import "fmt"

// ------------------------------------------------------------------- //
// Collection of error types for route and controller error handling.
// Author(s): 190014935
// ------------------------------------------------------------------- //

// Handle non-registered users.
type BadUserError struct {
	userID string
}

func (e *BadUserError) Error() string { return "User " + e.userID + " doesn't exist!" }

// Handle users with incorrect permissions.
type WrongPermissionsError struct {
	userID string
}

func (e *WrongPermissionsError) Error() string {
	return "User" + e.userID + "does not have required permissions!"
}

// Handle non-existant submissions.
type NoSubmissionError struct {
	ID uint
}

func (e *NoSubmissionError) Error() string { return fmt.Sprintf("Submission %d doesn't exist!", e.ID) }
