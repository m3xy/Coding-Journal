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

	ENDPOINT_TOKEN  = "/token"
	SUBROUTE_AUTH   = "/auth"
	ENDPOINT_LOGIN  = "/login"
	ENDPOINT_SIGNUP = "/register"
)

var JwtSecret string // JWT Secret variable.

// Subrouter
func getAuthSubRoutes(r *mux.Router) {
	auth := r.PathPrefix(SUBROUTE_AUTH).Subrouter()

	auth.HandleFunc(ENDPOINT_LOGIN, PostAuthLogIn).Methods(http.MethodPost, http.MethodOptions)
	auth.HandleFunc(ENDPOINT_SIGNUP, signUp).Methods(http.MethodPost, http.MethodOptions)

	// Set up jwt secret
	myEnv, err := godotenv.Read("secrets.env")
	if err == nil {
		JwtSecret = myEnv["JWT_SECRET"]
	}
}

// -- Log In -- //

/*
  Client-made log in endpoint
  Content-type: application/json
  Input: {"email": string, "password": string, "groupNumber": int}
  Success: 200, Credentials correc
  		Returns: { access_token, refresh_token, redirect-url, expires }
  Failure: 401, Unauthorized
  		Returns: { message: string, error: bool }
*/
func PostAuthLogIn(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received Auth Log-In request from %s.", r.RemoteAddr)

	// Check if the body's correctly formed.
	var credentials AuthLoginPostBody
	var encodable interface{}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		encodable = StandardResponse{
			Message: "Email or password incorrectly formed!",
			Error:   true,
		}
		log.Printf("[WARN] Request body not correctly formed: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		// Get access token from controller.
		response, status := ControllerAuthLogin(credentials)
		switch {
		case status == http.StatusInternalServerError:
			encodable = StandardResponse{Message: "Internal Server Error", Error: true}
			break
		case status == http.StatusUnauthorized:
			encodable = StandardResponse{Message: "Email or password incorrect!", Error: true}
			break
		case status == http.StatusOK:
			encodable = response
		}

		// Send response
		w.WriteHeader(status)
	}

	if err := json.NewEncoder(w).Encode(encodable); err != nil {
		log.Printf("[ERROR] Response encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

}

// Controller for the Auth Login POST method.
func ControllerAuthLogin(credentials AuthLoginPostBody) (AuthLogInResponse, int) {
	// Get where to fetch user from given group number.
	var uuid string
	if credentials.GroupNumber == TEAM_ID {
		userID, status := GetLocalUserID(JournalLoginPostBody{Email: credentials.Email, Password: credentials.Password})
		if status != http.StatusOK {
			return AuthLogInResponse{}, status
		}
		uuid = userID
	} else if userID, err := GetForeignUserID(credentials); err != nil {
		return AuthLogInResponse{}, http.StatusUnauthorized
	} else {
		uuid = userID
	}

	// Get tokens from successful login.
	var err error
	resp := AuthLogInResponse{
		RedirectUrl: "/user",
		Expires:     3600,
	}
	if resp.AccessToken, err = createToken(uuid, "bearer"); err != nil {
		return AuthLogInResponse{}, http.StatusInternalServerError
	}
	if resp.RefreshToken, err = createToken(uuid, "refresh"); err != nil {
		return AuthLogInResponse{}, http.StatusInternalServerError
	}
	return resp, http.StatusOK
}

// Get a foreign user's ID from a foreign journal's POST /login endpoint.
func GetForeignUserID(credentials AuthLoginPostBody) (string, error) {
	// Get foreign server to request login from.
	var retServer Server
	res := gormDb.Limit(1).Find(&retServer, credentials.GroupNumber)
	switch {
	case res.Error != nil:
		log.Printf("[ERROR] SQL query error: %v", res.Error)
		return "", res.Error
	case res.RowsAffected == 0:
		log.Printf("[WARN] Group number %d doesn't exist in database.", credentials.GroupNumber)
		return "", res.Error
	}

	// Send request to foreign server.
	var ResBody struct {
		ID string `json:"userId"`
	}
	if resp, err := func() (*http.Response, error) {
		if body, err := json.Marshal(JournalLoginPostBody{Email: credentials.Email, Password: credentials.Password}); err != nil {
			log.Printf("[ERROR] JSON body encoding failed: %v", err)
			return nil, err
		} else if req, err := http.NewRequest(http.MethodPost, retServer.Url+ENDPOINT_LOGIN, bytes.NewBuffer(body)); err != nil {
			log.Printf("[ERROR] Request creation failed: %v", err)
			return nil, err
		} else {
			req.Header.Set(SECURITY_TOKEN_KEY, retServer.Token)
			client := &http.Client{}
			return client.Do(req)
		}
	}(); err != nil {
		log.Printf("[WARN] Response getter failed: %v", err)
		return "", err // Error in request/response
	} else if err := json.NewDecoder(resp.Body).Decode(&ResBody); err != nil {
		log.Printf("[ERROR] Body decoding failed: %v", err)
		return "", err // Decoding error.
	} else {
		return ResBody.ID, nil
	}
}

// Get a local user ID's from the given credentials.
func GetLocalUserID(credentials JournalLoginPostBody) (string, int) {
	var user struct {
		GlobalUserID string
		Email        string
		Password     string
	}
	res := gormDb.Model(&User{}).Limit(1).Find(&user, "Email = ?", credentials.Email)
	switch {
	case res.Error != nil:
		log.Printf("[ERROR] SQL query error: %v", res.Error)
		return "", http.StatusInternalServerError
	case res.RowsAffected == 0:
		log.Printf("[WARN] User not found: %s", credentials.Email)
		return "", http.StatusUnauthorized
	case credentials.Email != user.Email || !comparePw(credentials.Password, user.Password):
		log.Printf("[WARN] User's password invalid: %s", user.GlobalUserID)
		return "", http.StatusUnauthorized
	}
	return user.GlobalUserID, http.StatusOK
}

// -- Token Control -- //

/*
  Client refresh token getter function.
  Content-type: application/json
  Input: {"Refresh": { "type": "string", "description": "User's valid refresh token."}}
*/
func GetToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("refresh_token")
	var body interface{}

	// Validate refresh token, and create new tokens.
	if ok, user := validateWebToken(refreshToken, "refresh"); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		body = StandardResponse{
			Message: "Given refresh token is invalid!",
			Error:   true,
		}
	} else if token, err := createToken(user, "bearer"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body = StandardResponse{
			Message: "Access token creation failed!",
			Error:   true,
		}
	} else if newRefresh, err := createToken(user, "refresh"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body = StandardResponse{
			Message: "Refresh token creation failed!",
			Error:   true,
		}
	} else {
		body = AuthLogInResponse{
			AccessToken:  token,
			RefreshToken: newRefresh,
			Expires:      3600,
			RedirectUrl:  "",
		}
	}

	// Send response
	if err := json.NewEncoder(w).Encode(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("[ERROR] Response parsing failed: %v", err)
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

// Create a token with given scope.
func createToken(ID string, scope string) (string, error) {
	claims := JwtClaims{
		ID:    ID,
		Scope: scope,
		StandardClaims: jwt.StandardClaims{
			Issuer: "CS3099User11_Project_Code",
		},
	}
	switch scope {
	case "refresh":
		claims.ExpiresAt = time.Now().Unix() + 72000
		break
	default:
		claims.ExpiresAt = time.Now().Unix() + 3600
		break
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte(JwtSecret))
}

// -- Sign Up -- //

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
	var message string
	user := &User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		log.Printf("[ERROR] JSON decoding failed: %v", err)
		message = "Registration failed - Wrong fields provided."
		goto ERROR
	}
	if validator.Validate(*user) != nil {
		log.Printf("[WARN] Invalid password format received.")
		message = "Registration failed - invalid password"
		goto ERROR
	}

	if _, err := registerUser(*user); err != nil {
		log.Printf("[ERROR] User registration failed: %v", err)
		message = "Registration failed - " + err.Error()
		goto ERROR
	}

	//
	log.Printf("[INFO] User signup from %s successful.", r.RemoteAddr)
	if err := json.NewEncoder(w).Encode(StandardResponse{
		Message: "Registration successful!",
		Error:   false,
	}); err != nil {
		log.Printf("[ERROR] JSON encoding failed: %v", err)
	}
	return

ERROR:
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(StandardResponse{
		Message: message,
		Error:   true,
	}); err != nil {
		log.Printf("[ERROR] JSON encoding failed: %v", err)
	}
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
