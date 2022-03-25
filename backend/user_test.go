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
		globalUsers[i].ID, _ = registerUser(testUsers[i], fmt.Sprint(i), fmt.Sprint(i), USERTYPE_NIL)
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
				(fmt.Sprint(i) == resCreds.FirstName) &&
				(fmt.Sprint(i) == resCreds.LastName) &&
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
			PhoneNumber: "07375942117",
			Organization: organization,
		}
		id, err := registerUser(user, fname, lname, userType)
		assert.NoError(t, err, "Error occurred while registering test user")
		return id
	}

	// wipe the db and filesystem submission tables
	clearUsers := func() {
		gormDb.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&User{})
		gormDb.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&GlobalUser{})
	}

	// handles sending the request to the specified endpoint
	handleQuery := func(queryRoute string) *http.Response {
		req, w := httptest.NewRequest(http.MethodGet, queryRoute, nil), httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Result()
	}

	t.Run("valid queries", func(t *testing.T) {
		defer clearUsers()
		userID1 := registerTestUser("test1@test.com", "Joe", "Shmo", USERTYPE_NIL, "org one")
		userID2 := registerTestUser("test2@test.com", "Bob", "Tao", USERTYPE_PUBLISHER, "org one two")
		userID3 := registerTestUser("test3@test.com", "Billy", "Tai", USERTYPE_REVIEWER, "org3")
		userID4 := registerTestUser("test4@test.com", "Will", "Zimmer", USERTYPE_EDITOR, "testtest")

		// handles sending the request and parsing the response for valid queries (i.e. status 200)
		handleValidQuery := func(queryRoute string) *QueryUsersResponse {
			resp := handleQuery(queryRoute)
			respData := &QueryUsersResponse{}
			if !assert.NoError(t, json.NewDecoder(resp.Body).Decode(respData), "Error decoding response body") {
				return nil
			} else if !assert.Falsef(t, respData.StandardResponse.Error,
				"Error returned on query - %v", respData.StandardResponse.Message) {
				return nil
			}
			return respData
		}

		// confirms that the query requests return a full user profile
		t.Run("confirm data", func(t *testing.T) {
			queryRoute := fmt.Sprintf("%s%s", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
			respData := handleValidQuery(queryRoute)
			user := respData.Users[0]
			switch {
			case !assert.NotEmpty(t, user.User, "user profile is nil"),
				!assert.NotEmpty(t, user.User.Email, "email is nil"),
				!assert.Empty(t, user.User.Password, "password returned!"),
				!assert.NotEmpty(t, user.FirstName, "first name is nil"),
				!assert.NotEmpty(t, user.LastName, "last name is nil"),
				!assert.NotEmpty(t, user.User.PhoneNumber, "phone number is nil"),
				!assert.NotEmpty(t, user.User.Organization, "organization is nil"):
				return
			}
		})

		t.Run("order by name", func(t *testing.T) {
			t.Run("first name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?orderBy=firstName", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
				respData := handleValidQuery(queryRoute)
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
				respData := handleValidQuery(queryRoute)
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
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID1, "incorrect user")
			})
			t.Run("publisher", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY_USER, USERTYPE_PUBLISHER)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID2, "incorrect user")
			})
			t.Run("reviewer", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY_USER, USERTYPE_REVIEWER)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID3, "incorrect user")
			})
			t.Run("editor", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY_USER, USERTYPE_EDITOR)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID4, "incorrect user")
			})
		})

		t.Run("filter by name", func(t *testing.T) {
			t.Run("full name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?name=Joe+Shmo", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID1, "incorrect user")
			})

			t.Run("partial name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?name=Ta", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 2, len(respData.Users), "incorrect number of users returned")
				assert.Contains(t, []string{userID2, userID3}, respData.Users[0].ID, "incorrect user")
				assert.Contains(t, []string{userID2, userID3}, respData.Users[1].ID, "incorrect user")
			})
		})

		t.Run("filter by organization", func(t *testing.T) {
			t.Run("full org", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?organization=testtest", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, userID4, respData.Users[0].ID, "incorrect user")
			})

			t.Run("partial org", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?organization=org", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
				respData := handleValidQuery(queryRoute)
				expectedUsers := []string{userID1, userID2, userID3}
				switch {
				case !assert.Equal(t, 3, len(respData.Users), "incorrect number of users returned"),
					!assert.Contains(t, expectedUsers, respData.Users[0].ID, "incorrect user"),
					!assert.Contains(t, expectedUsers, respData.Users[1].ID, "incorrect user"),
					!assert.Contains(t, expectedUsers, respData.Users[2].ID, "incorrect user"):
					return
				}
			})
		})

	})

	t.Run("request validation", func(t *testing.T) {
		// tests user types which are non-integers, or out of the 0-4 range
		// returns status 400 - bad request
		t.Run("userType", func(t *testing.T) {
			queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY_USER, -1)
			resp1 := handleQuery(queryRoute)
			queryRoute = fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY_USER, 5)
			resp2 := handleQuery(queryRoute)
			queryRoute = fmt.Sprintf("%s%s?userType=%f", SUBROUTE_USERS, ENDPOINT_QUERY_USER, 1.5)
			resp3 := handleQuery(queryRoute)
			switch {
			case !assert.Equal(t, http.StatusBadRequest, resp1.StatusCode, "incorrect status code"),
				!assert.Equal(t, http.StatusBadRequest, resp2.StatusCode, "incorrect status code"),
				!assert.Equal(t, http.StatusBadRequest, resp3.StatusCode, "incorrect status code"):
				return
			}
		})

		// tests user types which are non-integers, or out of the 0-4 range
		// returns status 400 - bad request
		t.Run("orderBy", func(t *testing.T) {
			queryRoute := fmt.Sprintf("%s%s?orderBy=invalid", SUBROUTE_USERS, ENDPOINT_QUERY_USER)
			resp := handleQuery(queryRoute)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "incorrect status code")
		})
	})
}