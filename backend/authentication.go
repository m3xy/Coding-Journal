// === === === === === === === === === === === === ===
// authentication.go
// Set of all functions relating to user authentication,
// registration, and migration.
//
// Authors: 190014935
// Creation Date: 19/10/2021
// Last Modified: 04/11/2021
// === === === === === === === === === === === === ===

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

const (
	SPECIAL_CHARS = "//!//@//#//$//%//^//&//*//,//.//;//://_//-//+//-//=//\"//'"
	A_NUMS        = "a-zA-Z0-9"
	CLAIM_BEARER  = "bearer"
	CLAIM_REFRESH = "refresh"
)

var JwtSecret string // JWT Secret variable.

// Subrouter
func getAuthSubRoutes(r *mux.Router) {
	r.Use(jwtMiddleware)
	r.HandleFunc(ENDPOINT_LOGIN, authLogIn).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc(ENDPOINT_SIGNUP, signUp).Methods(http.MethodPost, http.MethodOptions)

	// Set up jwt secret
	myEnv, err := godotenv.Read("secrets.env")
	if err == nil {
		JwtSecret = myEnv["JWT_SECRET"]
	}
}

// Validate the user's access token.
func validateWebToken(accessToken string, scope string) (bool, string) {
	tokenSplit := strings.Split(accessToken, " ")
	if len(tokenSplit) != 2 || strings.ToLower(tokenSplit[0]) != scope {
		return false, "-"
	}
	token, err := jwt.ParseWithClaims(tokenSplit[1], &JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(JwtSecret), nil
	})
	if err != nil {
		return false, "-"
	}
	if claims, ok := token.Claims.(*JwtClaims); !ok {
		return false, "-"
	} else if claims.Scope != "scope" {
		return false, "-"
	} else {
		return true, claims.ID
	}

}

/*
	Log in to website with any server's database.
	Content type: application/json
	Input: {"email": string, "password": string, "groupNumber": int}
	Success: 200, Credentials are correct.
	Failure: 401, Unauthorized
	Returns: userId
*/
func logInGlobal(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received global login request from %s.", r.RemoteAddr)
	propsMap := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&propsMap)
	if err != nil {
		log.Printf("[WARN] Invalid security token received from %s.", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Query path from team ID.
	var retServer Server
	res := gormDb.Limit(1).Find(&retServer, propsMap[getJsonTag(&Server{}, "GroupNumber")])
	if res.RowsAffected == 0 {
		log.Printf("[WARN] Group number %s doesn't exist in database.", propsMap[getJsonTag(&Server{}, "GroupNumber")])
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if res.Error != nil {
		log.Printf("[ERROR] SQL query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Make request from given URL and security token
	jsonBody, err := json.Marshal(propsMap)
	if err != nil {
		log.Printf("[ERROR] JSON body encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	globalReq, _ := http.NewRequest(
		"POST", retServer.Url+"/login", bytes.NewBuffer(jsonBody))
	globalReq.Header.Set(SECURITY_TOKEN_KEY, retServer.Token)

	// Get response from login request.
	client := &http.Client{}
	globalReq.Header.Set(SECURITY_TOKEN_KEY, retServer.Token)
	foreignRes, err := client.Do(globalReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("[ERROR] HTTP Request error: %v", err)
		return
	} else if foreignRes.StatusCode != http.StatusOK {
		log.Printf("[WARN] Foreign server login request failed, mirroring...")
		w.WriteHeader(foreignRes.StatusCode)
		return
	}

	mirrorProps := make(map[string]string)
	err = json.NewDecoder(foreignRes.Body).Decode(&mirrorProps)
	if err != nil {
		log.Printf("[ERROR] JSON decoding error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(&propsMap)
	if err != nil {
		log.Printf("[ERROR] JSON encoding error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

/*
  Client-made log in endpoint
  Content-type: application/json
  Input: {"email": string, "password": string, }
  Success: 200, Credentials correc
  		Returns: { access_token, refresh_token, redirect-url, expires }
  Failure: 401, Unauthorized
  		Returns: { message: string, error: bool }
*/
func authLogIn(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received auth log in request from %s.", r.RemoteAddr)

	var credentials struct {
		Email    string `json:"email" validate:"validEmail"`
		Password string `json:"password"`
	}
	// Error on incorrectly formed request.
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(&StandardResponse{
			Message: "Email or password incorrectly formed!",
			Error:   true,
		})
		return
	}

	// Validate user credentials
	var user struct {
		GlobalUserID string
		Email        string
		Password     string
	}
	if res := gormDb.Model(&User{}).Limit(1).Find(&user, "Email = ?", credentials.Email); res.Error != nil {
		log.Printf("[ERROR] SQL query error: %v\n", res.Error)
		goto INTERNAL
	} else if res.RowsAffected == 0 {
		log.Printf("[INFO] No user found!")
		goto CREDS
	} else if !comparePw(credentials.Password, user.Password) {
		log.Printf("[INFO] Password incorrect!")
		goto CREDS
	}

	// Send successful response
	if resp, err := createTokenSuite(user); err != nil {
		log.Printf("[ERROR] Access Token Creation Error: %v", err)
		goto INTERNAL
	} else if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] Response encoding error: %v", err)
		goto INTERNAL
	} else {
		w.WriteHeader(http.StatusOK)
		return
	}

CREDS:
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(&StandardResponse{
		Message: "Email or password incorrect!",
		Error:   true,
	})
	return
INTERNAL:
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(&StandardResponse{
		Message: "Internal Server Error",
		Error:   true,
	})
	return
}

// Create access and refresh tokens for the user.
func createTokenSuite(u struct {
	GlobalUserID string
	Email        string
	Password     string
}) (AuthLogInResponse, error) {
	access_claims := JwtClaims{
		ID:    u.GlobalUserID,
		Scope: "bearer",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + 3600,
			Issuer:    "CS3099User11_Project_Code",
		},
	}
	refresh_claims := JwtClaims{
		ID:    u.GlobalUserID,
		Scope: "refresh",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + 72000,
			Issuer:    "CS3099User11_Project_Code",
		},
	}
	resp := AuthLogInResponse{
		RedirectUrl: "/user",
		Expires:     3600,
	}

	access, refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, access_claims),
		jwt.NewWithClaims(jwt.SigningMethodHS512, refresh_claims)
	if signedAccess, err := access.SignedString([]byte(JwtSecret)); err != nil {
		return AuthLogInResponse{}, err
	} else {
		resp.AccessToken = signedAccess
	}
	if signedRefresh, err := refresh.SignedString([]byte(JwtSecret)); err != nil {
		return AuthLogInResponse{}, err
	} else {
		resp.RefreshToken = signedRefresh
	}

	return resp, nil
}

/*
  Router function to sign up to website.
  Content type: application/json
  Success: 200, OK
  Failure: 400, bad request
*/
func signUp(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received sign up request from %s.", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from JSON request and validate them.
	user := &User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		log.Printf("[ERROR] JSON decoding failed: %v", err)
		goto ERROR
	}
	if validator.Validate(*user) != nil {
		log.Printf("[WARN] Invalid password format received.")
		goto ERROR
	}

	if _, err := registerUser(*user); err != nil {
		log.Printf("[ERROR] User registration failed: %v", err)
		goto ERROR
	}
	log.Printf("[INFO] User signup from %s successful.", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	return

ERROR:
	w.WriteHeader(http.StatusBadRequest)
	return
}

// Register a user to the database. Returns user global ID.
func registerUser(user User) (string, error) {
	// Hash password and store new credentials to database.
	user.Password = string(hashPw(user.Password))

	registeredUser := GlobalUser{
		FullName: user.FirstName + " " + user.LastName,
		User:     user,
	}
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// Check constraints on user
		if !isUnique(tx, User{}, "Email", user.Email) {
			return errors.New("Email already taken!")
		}

		// Make credentials insert transaction.
		if err := gormDb.Create(&registeredUser).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return "", err
	}

	// Return user's primary key (the UUID)
	return registeredUser.ID, nil
}
