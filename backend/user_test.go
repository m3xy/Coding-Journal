package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Test user info getter.
func TestGetUserProfile(t *testing.T) {
	testInit()
	defer testEnd()

	// Start mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_USER+"/{id}", getUserProfile)

	// Populate database for testing and test valid user.
	var err error
	globalUsers := make([]GlobalUser, len(testGlobUsers))
	for i, u := range testGlobUsers {
		globalUsers[i] = *u.getCopy()
		globalUsers[i].ID, err = registerUser(globalUsers[i])
		if !assert.NoError(t, err, "error registering test users") {
			return
		}
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

			// tests for user equality
			switch {
			case !assert.Equal(t, globalUsers[i].FirstName, resCreds.FirstName, "first names not equal"),
				!assert.Equal(t, globalUsers[i].LastName, resCreds.LastName, "last names not equal"),
				!assert.Equal(t, globalUsers[i].User.Email, resCreds.User.Email, "emails not equal"),
				!assert.Equal(t, globalUsers[i].User.Organization, resCreds.User.Organization, "organizations not equal"),
				!assert.Equal(t, globalUsers[i].User.PhoneNumber, resCreds.User.PhoneNumber, "phone numbers not equal"):
				return
			}
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

func TestChangeUserPermissions(t *testing.T) {
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_USER+"/{id}"+ENDPOINT_CHANGE_PERMISSIONS, PostChangePermissions)

	testUser := *testGlobUsers[0].getCopy()
	userID, err := registerUser(testUser)
	if !assert.NoError(t, err, "Error occurred while registering test user") {
		return
	}

	testEditor := *testEditors[0].getCopy()
	editorID, err := registerUser(testEditor)
	if !assert.NoError(t, err, "Error occurred while registering test editor") {
		return
	}

	// sends query to the proper endpoint and returns the response
	handleQuery := func(queryRoute string, reqStruct ChangePermissionsPostBody, ctx *RequestContext) *http.Response {
		reqBody, err := json.Marshal(reqStruct)
		if !assert.NoError(t, err, "Error while marshalling assign reviewers body!") {
			return nil
		}
		req := httptest.NewRequest(http.MethodPost, queryRoute, bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()
		rCtx := context.WithValue(req.Context(), "data", ctx)
		router.ServeHTTP(w, req.WithContext(rCtx))
		return w.Result()
	}

	t.Run("valid request", func(t *testing.T) {
		queryRoute := fmt.Sprintf("%s/%s%s", SUBROUTE_USER, userID, ENDPOINT_CHANGE_PERMISSIONS)
		reqBody := ChangePermissionsPostBody{Permissions: USERTYPE_PUBLISHER}
		ctx := &RequestContext{ID: editorID, UserType: USERTYPE_EDITOR}
		resp := handleQuery(queryRoute, reqBody, ctx)
		queriedUser := &GlobalUser{}
		switch {
		case !assert.Equalf(t, http.StatusOK, resp.StatusCode, "incorrect status code"),
			!assert.NoError(t, gormDb.Select("user_type").Find(queriedUser, "id = ?", userID).Error, "user could not be retrieved"),
			!assert.Equal(t, USERTYPE_PUBLISHER, queriedUser.UserType, "permissions not changed"):
			return
		}
	})

	t.Run("request validation", func(t *testing.T) {

	})
}

func TestDeleteUser(t *testing.T) {
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_USER+"/{id}"+ENDPOINT_DELETE, PostDeleteUser)

	// registers a test users
	testUser := *testGlobUsers[0].getCopy()
	userID, err := registerUser(testUser)
	if !assert.NoError(t, err, "error registering user") {
		return
	}

	otherUser := *testGlobUsers[1].getCopy()
	otherUserID, err := registerUser(otherUser)
	if !assert.NoError(t, err, "error registering user") {
		return
	}

	handleQuery := func(queryRoute string, ctx *RequestContext) *http.Response {
		req, w := httptest.NewRequest(http.MethodGet, queryRoute, nil), httptest.NewRecorder()
		rCtx := context.WithValue(req.Context(), "data", ctx)
		router.ServeHTTP(w, req.WithContext(rCtx))
		return w.Result()
	}

	t.Run("delete non-logged in user", func(t *testing.T) {
		queryRoute := fmt.Sprintf("%s/%s%s", SUBROUTE_USER, userID, ENDPOINT_DELETE)
		ctx := &RequestContext{ID: otherUserID, UserType: otherUser.UserType}
		resp := handleQuery(queryRoute, ctx)
		if !assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "incorrect status code returned") {
			return
		}
	})

	t.Run("delete non-existant user", func(t *testing.T) {
		fakeID := "afeigrbnrilfdalkja-88ra9rakrb"
		queryRoute := fmt.Sprintf("%s/%s%s", SUBROUTE_USER, fakeID, ENDPOINT_DELETE)
		ctx := &RequestContext{ID: fakeID, UserType: USERTYPE_NIL}
		resp := handleQuery(queryRoute, ctx)
		if !assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "incorrect status code returned") {
			return
		}
	})

	t.Run("delete valid", func(t *testing.T) {
		queryRoute := fmt.Sprintf("%s/%s%s", SUBROUTE_USER, userID, ENDPOINT_DELETE)
		ctx := &RequestContext{ID: userID, UserType: testUser.UserType}
		resp := handleQuery(queryRoute, ctx)

		// checks that the user was deleted properly
		if !assert.Equal(t, http.StatusOK, resp.StatusCode, "incorrect status code returned") {
			return
		}
		res := gormDb.Find(&User{}, "global_user_id = ?", userID)
		if !assert.NoError(t, res.Error, "error querying db") {
			return
		}
		if !assert.Equal(t, int64(0), res.RowsAffected, "user not deleted!") {
			return
		}
		res = gormDb.Find(&GlobalUser{}, "id = ?", userID)
		if !assert.NoError(t, res.Error, "error querying db") {
			return
		}
		if !assert.Equal(t, int64(0), res.RowsAffected, "global user not deleted!") {
			return
		}
	})
}

func TestEditUser(t *testing.T) {
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_USER+"/{id}"+ENDPOINT_EDIT, PostEditUser)

	// registers a test users
	testUser := *testGlobUsers[0].getCopy()
	userID, err := registerUser(testUser)
	if !assert.NoError(t, err, "error registering user") {
		return
	}

	otherUser := *testGlobUsers[1].getCopy()
	otherUserID, err := registerUser(otherUser)
	if !assert.NoError(t, err, "error registering user") {
		return
	}

	handleQuery := func(queryRoute string, ctx *RequestContext, reqStruct *EditUserPostBody) *http.Response {
		reqBody, err := json.Marshal(reqStruct)
		if !assert.NoError(t, err, "Error while marshalling assign reviewers body!") {
			return nil
		}
		req := httptest.NewRequest(http.MethodGet, queryRoute, bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()
		rCtx := context.WithValue(req.Context(), "data", ctx)
		router.ServeHTTP(w, req.WithContext(rCtx))
		return w.Result()
	}

	// helper function to test whether the new user profile matches that which was given
	testEdit := func(newProfile *EditUserPostBody, userID string) {
		// gets the user objects
		globUser := &GlobalUser{}
		if !assert.NoError(t, gormDb.Model(globUser).Preload("User").Find(globUser, "id = ?", userID).Error,
			"could not retrieve user") {
			return
		}
		// tests for equality between the queried user and the new profile (only testing for equality in non-empty fields)
		if len(newProfile.Password) > 0 &&
			!assert.True(t, comparePw(newProfile.Password, globUser.User.Password), "passwords do not match") {
			return
		} else if len(newProfile.FirstName) > 0 &&
			!assert.Equal(t, newProfile.FirstName, globUser.FirstName, "first names do not match") {
			return
		} else if len(newProfile.LastName) > 0 &&
			!assert.Equal(t, newProfile.LastName, globUser.LastName, "last names do not match") {
			return
		} else if len(newProfile.PhoneNumber) > 0 &&
			!assert.Equal(t, newProfile.PhoneNumber, globUser.User.PhoneNumber, "phone numbers do not match") {
			return
		} else if len(newProfile.Organization) > 0 &&
			!assert.Equal(t, newProfile.Organization, globUser.User.Organization, "organizations do not match") {
			return
		}
	}

	t.Run("valid cases", func(t *testing.T) {
		handleValidEdit := func(reqStruct *EditUserPostBody) {
			queryRoute := fmt.Sprintf("%s/%s%s", SUBROUTE_USER, userID, ENDPOINT_EDIT)
			ctx := &RequestContext{ID: userID, UserType: otherUser.UserType}
			resp := handleQuery(queryRoute, ctx, reqStruct)
			if !assert.Equal(t, http.StatusOK, resp.StatusCode, "incorrect status code returned") {
				return
			}
			testEdit(reqStruct, userID)
		}
		t.Run("edit password", func(t *testing.T) {
			handleValidEdit(&EditUserPostBody{Password: VALID_PW + "a"})
		})
		t.Run("edit first name", func(t *testing.T) {
			handleValidEdit(&EditUserPostBody{FirstName: "newName"})
		})
		t.Run("edit last name", func(t *testing.T) {
			handleValidEdit(&EditUserPostBody{LastName: "newName"})
		})
		t.Run("edit phone number", func(t *testing.T) {
			handleValidEdit(&EditUserPostBody{PhoneNumber: "07375942117"})
		})
		t.Run("edit organization", func(t *testing.T) {
			handleValidEdit(&EditUserPostBody{Organization: "org org org"})
		})
	})

	t.Run("request validation", func(t *testing.T) {
		t.Run("edit non-logged in user", func(t *testing.T) {
			queryRoute := fmt.Sprintf("%s/%s%s", SUBROUTE_USER, userID, ENDPOINT_EDIT)
			ctx := &RequestContext{ID: otherUserID, UserType: otherUser.UserType}
			reqStruct := &EditUserPostBody{}
			resp := handleQuery(queryRoute, ctx, reqStruct)
			if !assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "incorrect status code returned") {
				return
			}
		})

		t.Run("delete non-existant user", func(t *testing.T) {
			fakeID := "afeigrbnrilfdalkja-88ra9rakrb"
			queryRoute := fmt.Sprintf("%s/%s%s", SUBROUTE_USER, fakeID, ENDPOINT_EDIT)
			ctx := &RequestContext{ID: fakeID, UserType: USERTYPE_NIL}
			reqStruct := &EditUserPostBody{}
			resp := handleQuery(queryRoute, ctx, reqStruct)
			if !assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "incorrect status code returned") {
				return
			}
		})
	})
}

func TestGetUserQuery(t *testing.T) {
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_USERS+ENDPOINT_QUERY, GetQueryUsers)

	// registers a test user with the given fields and returns their global ID (just makes calling registerUser more compact)
	registerUser := func(email string, fname string, lname string, userType int, organization string) string {
		user := User{
			Email:        email,
			Password:     VALID_PW,
			PhoneNumber:  "07375942117",
			Organization: organization,
		}
		globUser := GlobalUser{FirstName: fname, LastName: lname, UserType: userType, User: &user}
		id, err := registerUser(globUser)
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
		userID1 := registerUser("test1@test.com", "Joe", "Shmo", USERTYPE_NIL, "org one")
		userID2 := registerUser("test2@test.com", "Bob", "Tao", USERTYPE_PUBLISHER, "org one two")
		userID3 := registerUser("test3@test.com", "Billy", "Tai", USERTYPE_REVIEWER, "org3")
		userID4 := registerUser("test4@test.com", "Will", "Zimmer", USERTYPE_EDITOR, "testtest")

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

		// confirms that the query requests return a full user profile not just ID
		t.Run("confirm data", func(t *testing.T) {
			queryRoute := fmt.Sprintf("%s%s", SUBROUTE_USERS, ENDPOINT_QUERY)
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
				queryRoute := fmt.Sprintf("%s%s?orderBy=firstName", SUBROUTE_USERS, ENDPOINT_QUERY)
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
				queryRoute := fmt.Sprintf("%s%s?orderBy=lastName", SUBROUTE_USERS, ENDPOINT_QUERY)
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
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY, USERTYPE_NIL)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID1, "incorrect user")
			})
			t.Run("publisher", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY, USERTYPE_PUBLISHER)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID2, "incorrect user")
			})
			t.Run("reviewer", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY, USERTYPE_REVIEWER)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID3, "incorrect user")
			})
			t.Run("editor", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY, USERTYPE_EDITOR)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID4, "incorrect user")
			})
		})

		t.Run("filter by name", func(t *testing.T) {
			t.Run("full name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?name=Joe+Shmo", SUBROUTE_USERS, ENDPOINT_QUERY)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, respData.Users[0].ID, userID1, "incorrect user")
			})

			t.Run("partial name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?name=Ta", SUBROUTE_USERS, ENDPOINT_QUERY)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 2, len(respData.Users), "incorrect number of users returned")
				assert.Contains(t, []string{userID2, userID3}, respData.Users[0].ID, "incorrect user")
				assert.Contains(t, []string{userID2, userID3}, respData.Users[1].ID, "incorrect user")
			})
		})

		t.Run("filter by organization", func(t *testing.T) {
			t.Run("full org", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?organization=testtest", SUBROUTE_USERS, ENDPOINT_QUERY)
				respData := handleValidQuery(queryRoute)
				assert.Equal(t, 1, len(respData.Users), "incorrect number of users returned")
				assert.Equal(t, userID4, respData.Users[0].ID, "incorrect user")
			})

			t.Run("partial org", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?organization=org", SUBROUTE_USERS, ENDPOINT_QUERY)
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
			queryRoute := fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY, -1)
			resp1 := handleQuery(queryRoute)
			queryRoute = fmt.Sprintf("%s%s?userType=%d", SUBROUTE_USERS, ENDPOINT_QUERY, 5)
			resp2 := handleQuery(queryRoute)
			queryRoute = fmt.Sprintf("%s%s?userType=%f", SUBROUTE_USERS, ENDPOINT_QUERY, 1.5)
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
			queryRoute := fmt.Sprintf("%s%s?orderBy=invalid", SUBROUTE_USERS, ENDPOINT_QUERY)
			resp := handleQuery(queryRoute)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "incorrect status code")
		})
	})
}
