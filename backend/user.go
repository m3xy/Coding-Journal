package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

const (
	SUBROUTE_USERS      = "/users"
	SUBROUTE_USER       = "/user"

	ENDPOINT_GET        = "/get"
	ENDPOINT_CHANGE_PERMISSIONS = "/changepermissions"
	ENDPOINT_QUERY_USER = "/query"
	ENDPOINT_DELETE_USER = "/delete"
)

func getUserSubroutes(r *mux.Router) {
	user := r.PathPrefix(SUBROUTE_USER).Subrouter()
	users := r.PathPrefix(SUBROUTE_USERS).Subrouter()

	users.Use(jwtMiddleware)

	// User routes:
	// + GET /user/{id} - Get given user profile.
	// + POST /user/{id}/changepermissions - editor changing user permissions
	user.HandleFunc("/{id}", getUserProfile).Methods(http.MethodGet)
	user.HandleFunc("/{id}"+ENDPOINT_CHANGE_PERMISSIONS, PostChangePermissions).Methods(http.MethodOptions, http.MethodPost)

	// Users routes:
	// + GET /users/query
	users.HandleFunc(ENDPOINT_QUERY_USER, GetQueryUsers).Methods(http.MethodGet)
}

func getUserOutFromUser(tx *gorm.DB) *gorm.DB {
	return tx.Select("GlobalUserID", "Email", "PhoneNumber", "Organization", "CreatedAt")
}

// -----------
// Router functions
// -----------

/*
	Get user profile info for a user.
	Content type: application/json
	Success: 200, Credentials can be passed down.
	Failure: 404, User not found.
*/
func getUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get user details from user ID.
	vars := mux.Vars(r)
	user := &GlobalUser{ID: vars["id"]}
	if res := gormDb.Preload("AuthoredSubmissions").Preload("User", getUserOutFromUser).Limit(1).Find(&user); res.Error != nil {
		log.Printf("[ERROR] SQL query error: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if res.RowsAffected == 0 {
		log.Printf("[WARN] No user linked to %s", vars["id"])
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Encode user and send.
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("[ERROR] User data JSON encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Router function for editors to change user's permissions
// POST /user/{id}/changepermissions
func PostChangePermissions(w http.ResponseWriter, r *http.Request) {
	reqBody := &ChangePermissionsPostBody{}
	resp := &StandardResponse{Message: "successfully changed permissions!", Error: false}
	// gets the user ID from the vars and logged in user details from request context
	params := mux.Vars(r)
	userID := string(params["id"])
	if ctx, ok := r.Context().Value("data").(*RequestContext); !ok || validate.Struct(ctx) != nil {
		resp = &StandardResponse{Message: "Request Context not set, user not logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if ctx.UserType != USERTYPE_EDITOR { // logged in user must be an editor to change another user's permissions
		resp = &StandardResponse{Message: 
			"The client must have editor permissions to change another user's permissions.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil || validate.Struct(reqBody) != nil {
		// request body could not be validated or decoded
		resp = &StandardResponse{Message: "Unable to parse request body.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	} else if err := ControllerUpdatePermissions(userID, reqBody.Permissions); err != nil {
		switch err.(type) {
		case *BadUserError:
			resp = &StandardResponse{Message:"Cannot update permissions - user does not exist", Error:true}
			w.WriteHeader(http.StatusBadRequest)

		// Unexpected error - error out as server error.
		default:
			log.Printf("[ERROR] could not change user permissions: %v\n", err)
			resp = &StandardResponse{Message: "Internal Server Error - could not change user permissions", Error: true}
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// Return response body after function successful.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Controller to change a user's permissions (UserType)
func ControllerUpdatePermissions(userID string, permissions int) error {
	if res := gormDb.Model(&GlobalUser{ID: userID}).
		Update("user_type", permissions); res.Error != nil {
			return res.Error
	} else if res.RowsAffected == 0 {
		return &BadUserError{userID: userID}
	}
	return nil
}

// router function for users to delete their accounts
// POST /user/{id}/delete
func PostDeleteUser(w http.ResponseWriter, r *http.Request) {
	resp := StandardResponse{ Message: "user deleted successfully", Error: false }

	// parse the userID from the url
	params := mux.Vars(r)
	userID := string(params["id"])

	// validates request context 
	if ctx, ok := r.Context().Value("data").(*RequestContext); !ok || validate.Struct(ctx) != nil {
		resp = StandardResponse{ Message: "Bad Request Context"}
		w.WriteHeader(http.StatusBadRequest)
	
	} else if userID != ctx.ID {
		resp = StandardResponse{ Message: "Unauthorized - cannot delete an account which is not yours", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if err := ControllerDeleteUser(ctx.ID); err != nil {
		switch err.(type) {
		case *BadUserError:
			resp = StandardResponse{ Message: fmt.Sprintf("Bad Request - User %s does not exist!", ctx.ID), Error: true}
			w.WriteHeader(http.StatusBadRequest)
		default:
			log.Printf("[ERROR] could not delete user %s: %v", userID, err)
			resp = StandardResponse{ Message: "Internal Server Error - could not delete user", Error: true }
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// writes the response
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting json response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Controller function to do the actual work of removing a user from the database
func ControllerDeleteUser(userID string) error {
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// deletes global user
		globUser := &GlobalUser{ID: userID}
		if res := gormDb.Delete(globUser); res.Error != nil {
			return res.Error
		} else if res.RowsAffected != 1 {
			return &BadUserError{ userID: userID }
		}
		// deletes local user
		if res := gormDb.Delete(&User{}, "global_user_id = ?", userID); res.Error != nil {
			return res.Error
		} else if res.RowsAffected != 1 {
			return &BadUserError{ userID: userID }
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}


// generalized query function for users
// GET /users/query
func GetQueryUsers(w http.ResponseWriter, r *http.Request) {
	var err error
	var stdResp StandardResponse
	var resp *QueryUsersResponse
	var users []GlobalUser

	// gets the request context if there is a user logged in (login not required, but if logged in, context must be valid)
	if ctx, ok := r.Context().Value("data").(*RequestContext); ok && validate.Struct(ctx) != nil {
		stdResp = StandardResponse{Message: "Bad Request Context", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	} else if users, err = ControllerQueryUsers(r.URL.Query(), ctx); err != nil {
		switch err.(type) {
		case *BadQueryParameterError:
			stdResp = StandardResponse{Message: fmt.Sprintf("Bad Request - %s", err.Error()), Error: true}
			w.WriteHeader(http.StatusBadRequest)
		case *ResultSetEmptyError:
			stdResp = StandardResponse{Message: "No submissions fit search queries", Error: false}
			w.WriteHeader(http.StatusOK)
		default:
			log.Printf("[ERROR] could not query users: %v\n", err)
			stdResp = StandardResponse{Message: "Internal Server Error - could not query users", Error: true}
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		stdResp = StandardResponse{Message: "", Error: false}
	}
	// builds the full response from the error message
	resp = &QueryUsersResponse{
		StandardResponse: stdResp,
		Users:            users,
	}

	// sends a response to the client
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Controller function to query the database for a list of users
func ControllerQueryUsers(queryParams url.Values, ctx *RequestContext) ([]GlobalUser, error) {
	var users []GlobalUser
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		tx = tx.Model(&GlobalUser{}).Joins("User")

		// filter results based on user type
		if len(queryParams["userType"]) > 0 {
			userType, err := strconv.Atoi(queryParams["userType"][0])
			if err != nil || userType < 0 || userType > 4 {
				return &BadQueryParameterError{ParamName: "userType", Value: queryParams["userType"][0]}
			}
			tx = tx.Where("user_type = ?", userType)
		}
		// filter by organization
		if len(queryParams["organization"]) > 0 {
			tx = filterByOrganization(tx, regexp.QuoteMeta(queryParams["organization"][0]))
		}
		// RegEx filtering for user name
		if len(queryParams["name"]) > 0 {
			tx = filterByUserName(tx, regexp.QuoteMeta(queryParams["name"][0]))
		}
		// order users alphabetically
		if len(queryParams["orderBy"]) > 0 {
			orderBy := queryParams["orderBy"][0]
			if orderBy != "firstName" && orderBy != "lastName" {
				return &BadQueryParameterError{ParamName: "orderBy", Value: queryParams["orderBy"]}
			}
			tx = orderUserQuery(tx, orderBy)
		}

		// executes full query
		if res := tx.Find(&users); res.Error != nil {
			return res.Error
		} else if res.RowsAffected == 0 {
			return &ResultSetEmptyError{}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	// removes passwords from the users
	for _, user := range users {
		user.User.Password = ""
	}
	return users, nil
}

// uses SQL REGEX to filter the users returned based on their names
func filterByUserName(tx *gorm.DB, userName string) *gorm.DB {
	params := map[string]interface{}{"full": userName}
	whereString := " global_users.first_name REGEXP @full OR global_users.last_name REGEXP @full"
	// only adds multiple regex conditions if the name given is multiple words
	if wordList := strings.Fields(userName); len(wordList) > 1 {
		for index, field := range wordList {
			whereString = whereString + fmt.Sprintf(
				" OR global_users.first_name REGEXP @%d OR global_users.last_name REGEXP @%d ", index, index)
			params[fmt.Sprint(index)] = field
		}
	}
	return tx.Where(whereString, params)
}

// uses SQL REGEX to filter the users returned based on their organization
func filterByOrganization(tx *gorm.DB, organization string) *gorm.DB {
	params := map[string]interface{}{"full": organization}
	whereString := " User.organization REGEXP @full"
	// only adds multiple regex conditions if the name given is multiple words
	if wordList := strings.Fields(organization); len(wordList) > 1 {
		for index, field := range wordList {
			whereString = whereString + fmt.Sprintf(" OR User.organization REGEXP @%d", index)
			params[fmt.Sprint(index)] = field
		}
	}
	return tx.Where(whereString, params)
}

// adds a piece to an sql query to order the results
func orderUserQuery(tx *gorm.DB, orderBy string) *gorm.DB {
	// order of submissions
	if orderBy == "lastName" {
		tx = tx.Order("global_users.last_name")
	} else if orderBy == "firstName" {
		tx = tx.Order("global_users.first_name")
	}
	return tx
}
