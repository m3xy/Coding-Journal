package main

import (
	"encoding/json"
	"log"
	"net/http"
	"fmt"
	"strings"
	"net/url"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

const (
	SUBROUTE_USERS = "/users"
	SUBROUTE_USER  = "/user"
	ENDPOINT_GET   = "/get"
	ENDPOINT_QUERY_USER = "/query"
)

func getUserSubroutes(r *mux.Router) {
	user := r.PathPrefix(SUBROUTE_USER).Subrouter()
	users := r.PathPrefix(SUBROUTE_USERS).Subrouter()

	users.Use(jwtMiddleware)

	// User routes:
	// + GET /user/{id} - Get given user profile.
	user.HandleFunc("/{id}", getUserProfile).Methods(http.MethodGet)

	// Users routes:
	// + GET /users/query
	users.HandleFunc(ENDPOINT_QUERY_USER, GetQueryUsers).Methods(http.MethodGet)
}

func getUserOutFromUser(tx *gorm.DB) *gorm.DB {
	return tx.Select("GlobalUserID", "Email", "FirstName", "LastName", "PhoneNumber", "Organization", "CreatedAt")
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

// generalized query function for users
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

func ControllerQueryUsers(queryParams url.Values, ctx *RequestContext) ([]GlobalUser, error) {
	var users []GlobalUser
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		tx = tx.Model(&GlobalUser{}).Joins("User")

		// filter results based on user type
		if len(queryParams["userType"]) > 0 {
			userType, err := strconv.Atoi(queryParams["userType"][0])
			if err != nil {
				return &BadQueryParameterError{ParamName:"userType", Value:queryParams["userType"][0]}
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
	return users, nil
}

// uses SQL REGEX to filter the users returned based on their names
func filterByUserName(tx *gorm.DB, userName string) *gorm.DB {
	params := map[string]interface{}{"full": userName}
	whereString := " User.first_name REGEXP @full OR User.last_name REGEXP @full"
	// only adds multiple regex conditions if the name given is multiple words
	if wordList := strings.Fields(userName); len(wordList) > 1 {
		for index, field := range wordList {
			whereString = whereString + fmt.Sprintf(" OR User.first_name REGEXP @%d OR User.last_name REGEXP @%d ", index, index)
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
		tx = tx.Order("User__last_name")
	} else if orderBy == "firstName" {
		tx = tx.Order("User__first_name")
	}
	return tx
}
