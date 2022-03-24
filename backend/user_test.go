package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"

	"gorm.io/gorm"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Test user info getter.
func TestGetUserProfile(t *testing.T) {
	testInit()
	defer testEnd()

	// Start mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_USER+"/{id}", getUserProfile)

	// Populate database for testing and test valid user.
	globalUsers := make([]GlobalUser, len(testUsers))
	for i := range testUsers {
		globalUsers[i].ID, _ = registerUser(testUsers[i], USERTYPE_NIL)
	}

	t.Run("Valid user profiles", func(t *testing.T) {
		for i, u := range globalUsers {
			r, w := httptest.NewRequest(http.MethodGet, SUBROUTE_USER+"/"+u.ID, nil), httptest.NewRecorder()
			router.ServeHTTP(w, r)
			res := w.Result()

			assert.Equal(t, http.StatusOK, res.StatusCode, "Status should be OK.")

			resCreds := GlobalUser{}
			if err := json.NewDecoder(res.Body).Decode(&resCreds); !assert.Nil(t, err, "JSON decoding must not error.") {
				return
			}

			// Check equality for all user info.
			equal := (testUsers[i].Email == resCreds.User.Email) &&
				(testUsers[i].FirstName == resCreds.User.FirstName) &&
				(testUsers[i].LastName == resCreds.User.LastName) &&
				(testUsers[i].PhoneNumber == resCreds.User.PhoneNumber) &&
				(testUsers[i].Organization == resCreds.User.Organization)
			assert.Equal(t, true, equal, "Users should be equal.")
		}
	})

	// Test invalid users.
	t.Run("Invalid user profile", func(t *testing.T) {
		r, w := httptest.NewRequest(http.MethodGet, SUBROUTE_USER+"/"+INVALID_ID, nil), httptest.NewRecorder()
		router.ServeHTTP(w, r)
		res := w.Result()

		assert.Equalf(t, http.StatusNotFound, res.StatusCode, "Request should return status %d", http.StatusNotFound)
	})
}

func TestGetUserQuery(t *testing.T) {
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_USERS+ENDPOINT_QUERY_USER, GetQueryUsers)

	// registers a test user with the given fields and returns their global ID
	registerTestUser := func(email string, fname string, lname string, userType int, organization string) string {
		user := User{
			Email: email,
			Password: VALID_PW, // in authentication_test.go
			FirstName: fname,
			LastName: lname,
			PhoneNumber: "07375942117",
			Organization: organization,
		}
		id, err := registerUser(user, userType)
		assert.NoError(t, err, "Error occurred while registering test user")
		return id
	}

	// wipe the db and filesystem submission tables
	clearUsers := func() {
		gormDb.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&User{})
		gormDb.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&GlobalUser{})
	}

	// handles sending the request and parsing the response
	handleQuery := func(queryRoute string) *QueryUsersResponse {
		req, w := httptest.NewRequest(http.MethodGet, queryRoute, nil), httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()

		respData := &QueryUsersResponse{}
		if !assert.NoError(t, json.NewDecoder(resp.Body).Decode(respData), "Error decoding response body") {
			return nil
		} else if !assert.Falsef(t, respData.StandardResponse.Error,
			"Error returned on query - %v", respData.StandardResponse.Message) {
			return nil
		}
		return respData
	}

	t.Run("valid queries", func(t *testing.T) {
		defer clearUsers()
		userID1 := registerTestUser("test1@test.com", "Joe", "Shmo", USERTYPE_NIL, "org1")
		userID2 := registerTestUser("test2@test.com", "Bob", "Tao", USERTYPE_PUBLISHER, "org2")
		userID3 := registerTestUser("test3@test.com", "Billy", "Tai", USERTYPE_REVIEWER, "org3")
		userID4 := registerTestUser("test4@test.com", "Will", "Zimmer", USERTYPE_EDITOR, "org4")

		t.Run("order by name", func(t *testing.T) {
			t.Run("first name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?orderBy=firstName", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
				respData := handleQuery(queryRoute)
				switch {
				case !assert.Equal(t, respData.Users[0].ID, userID3, "incorrect user order"),
					!assert.Equal(t, respData.Users[1].ID, userID2, "incorrect user order"),
					!assert.Equal(t, respData.Users[2].ID, userID1, "incorrect user order"),
					!assert.Equal(t, respData.Users[3].ID, userID4, "incorrect user order"):
					return
				}
			})
			t.Run("last name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?orderBy=lastName", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
				respData := handleQuery(queryRoute)
				switch {
				case !assert.Equal(t, respData.Users[0].ID, userID1, "incorrect user order"),
					!assert.Equal(t, respData.Users[1].ID, userID3, "incorrect user order"),
					!assert.Equal(t, respData.Users[2].ID, userID2, "incorrect user order"),
					!assert.Equal(t, respData.Users[3].ID, userID4, "incorrect user order"):
					return
				}
			})
		})

		t.Run("filter by permissions", func(t *testing.T) {
			t.Run("nil", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY_USER, USERTYPE_NIL)
				respData := handleQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID1, "incorrect user")
			})
			t.Run("publisher", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY_USER, USERTYPE_PUBLISHER)
				respData := handleQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID2, "incorrect user")
			})
			t.Run("reviewer", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY_USER, USERTYPE_REVIEWER)
				respData := handleQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID3, "incorrect user")
			})
			t.Run("editor", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY_USER, USERTYPE_EDITOR)
				respData := handleQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID4, "incorrect user")
			})
		})

		t.Run("filter by name", func(t *testing.T) {
			t.Run("full name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?name=Joe+Shmo", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
				respData := handleQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID1, "incorrect user")
			})

			t.Run("partial name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?name=Ta", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
				respData := handleQuery(queryRoute)
				assert.Equal(t, 2, len(respData.Users), "incorrect number of users returned")
				assert.Contains(t, []string{userID2, userID3}, respData.Users[0].ID, "incorrect user")
				assert.Contains(t, []string{userID2, userID3}, respData.Users[1].ID, "incorrect user")
			})
		})

	})

	t.Run("request validation", func(t *testing.T) {

	})
}