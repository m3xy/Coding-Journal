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
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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

type JwtClaims struct {
	ID       string `json:"userId"`
	UserType int    `json:"userType" validate:"min=0,max=4"`
	Scope    string
	jwt.StandardClaims
}

// Subrouter
func getAuthSubRoutes(r *mux.Router) {
	auth := r.PathPrefix(SUBROUTE_AUTH).Subrouter()

	// Authentication routes:
	// + POST /auth/login - Log in.
	// + POST /auth/register - Register.
	// + GET /auth/token  - Get new access token from a refresh token.
	auth.HandleFunc(ENDPOINT_LOGIN, PostAuthLogIn).Methods(http.MethodPost, http.MethodOptions)
	auth.HandleFunc(ENDPOINT_SIGNUP, PostSignUp).Methods(http.MethodPost, http.MethodOptions)
	auth.HandleFunc(ENDPOINT_TOKEN, GetToken).Methods(http.MethodGet)

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
	var permissions int
	if credentials.GroupNumber == TEAM_ID {
		userID, userType, status := GetLocalUserID(JournalLoginPostBody{Email: credentials.Email, Password: credentials.Password})
		if status != http.StatusOK {
			return AuthLogInResponse{}, status
		}
		uuid = userID
		permissions = userType
	} else if userID, err := GetForeignUserID(credentials); err != nil {
		return AuthLogInResponse{}, http.StatusUnauthorized
	} else {
		uuid = userID
		permissions = USERTYPE_NIL
	}

	// Get tokens from successful login.
	var err error
	resp := AuthLogInResponse{
		RedirectUrl: "/user",
		Expires:     3600,
	}
	if resp.AccessToken, err = createToken(uuid, permissions, "bearer"); err != nil {
		return AuthLogInResponse{}, http.StatusInternalServerError
	}
	if resp.RefreshToken, err = createToken(uuid, permissions, "refresh"); err != nil {
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
func GetLocalUserID(credentials JournalLoginPostBody) (string, int, int) {
	var user struct {
		GlobalUserID string
		Email        string
		Password     string
	}
	// struct to get user permissions
	var globalUser struct {
		UserType int
	}

	res := gormDb.Model(&User{}).Limit(1).Find(&user, "Email = ?", credentials.Email)
	switch {
	case res.Error != nil:
		log.Printf("[ERROR] SQL query error: %v", res.Error)
		return "", -1, http.StatusInternalServerError
	case res.RowsAffected == 0:
		log.Printf("[WARN] User not found: %s", credentials.Email)
		return "", -1, http.StatusUnauthorized
	case credentials.Email != user.Email || !comparePw(credentials.Password, user.Password):
		log.Printf("[WARN] User's password invalid: %s", user.GlobalUserID)
		return "", -1, http.StatusUnauthorized
	}
	// gets the user type provided a local user exists
	res = gormDb.Model(&GlobalUser{}).Limit(1).Find(&globalUser, "id = ?", user.GlobalUserID)
	if res.Error != nil {
		log.Printf("[ERROR] SQL query error: %v", res.Error)
		return "", -1, http.StatusInternalServerError
	}
	return user.GlobalUserID, globalUser.UserType, http.StatusOK
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
	if ok, user, userType := validateWebToken(refreshToken, CLAIM_REFRESH); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		body = StandardResponse{
			Message: "Given refresh token is invalid!",
			Error:   true,
		}
	} else if token, err := createToken(user, userType, CLAIM_BEARER); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body = StandardResponse{
			Message: "Access token creation failed!",
			Error:   true,
		}
	} else if newRefresh, err := createToken(user, userType, CLAIM_REFRESH); err != nil {
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
func validateWebToken(accessToken string, scope string) (bool, string, int) {
	tokenSplit := strings.Split(accessToken, " ")
	if len(tokenSplit) != 2 || strings.ToLower(tokenSplit[0]) != scope {
		return false, "-", -1
	}
	token, err := jwt.ParseWithClaims(tokenSplit[1], &JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(JwtSecret), nil
	})
	if err != nil {
		return false, "-", -1
	}
	if claims, ok := token.Claims.(*JwtClaims); !ok {
		return false, "-", -1
	} else if claims.Scope != scope {
		return false, "-", -1
	} else {
		return true, claims.ID, claims.UserType
	}
}

// Create a token with given scope.
func createToken(ID string, userType int, scope string) (string, error) {
	claims := JwtClaims{
		ID:       ID,
		UserType: userType,
		Scope:    scope,
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
func PostSignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from JSON request and validate them.
	user := &SignUpPostBody{}
	var resp FormResponse
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		log.Printf("[ERROR] JSON decoding failed: %v", err)
		resp = FormResponse{StandardResponse: StandardResponse{
			Message: "Invalid fields provided.",
			Error:   true,
		}}
	} else if validate.Struct(*user) != nil {
		log.Println("A")
		w.WriteHeader(http.StatusBadRequest)
		resp = FormResponse{StandardResponse: StandardResponse{
			Message: "Registration failed",
			Error:   true,
		}}
	} else if _, err := ControllerRegisterUser(*user, USERTYPE_REVIEWER_PUBLISHER); err != nil {
		log.Printf("[ERROR] User registration failed: %v", err)
		switch err.(type) {
		case *RepeatEmailError:
			w.WriteHeader(http.StatusBadRequest)
			resp = FormResponse{
				StandardResponse: StandardResponse{
					Message: err.Error(),
					Error:   true,
				},
				Fields: []struct {
					Field   string `json:"field"`
					Message string `json:"message"`
				}{{Field: "email", Message: err.Error()}},
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
			resp = FormResponse{StandardResponse: StandardResponse{
				Message: "Registration Failed, please try again later.",
				Error:   true,
			}}
		}
	} else {
		resp = FormResponse{StandardResponse: StandardResponse{
			Message: "Registration successful!",
			Error:   false,
		}}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] JSON encoding failed: %v", err)
	}
	return
}

// controller function to register a user
func ControllerRegisterUser(user SignUpPostBody, UserType int) (string, error) {
	// Hash password and store new credentials to database.
	user.Password = string(hashPw(user.Password))

	registeredUser := GlobalUser{
		FirstName: user.FirstName,
		LastName: user.LastName,
		UserType: UserType,
		User:     &User{
			Email: user.Email,
			Password: user.Password,
			PhoneNumber: user.PhoneNumber,
			Organization: user.Organization,
		},
	}
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// Check constraints on user
		if !isUnique(tx, User{}, "Email", user.Email) {
			return &RepeatEmailError{email: user.Email}
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

// -- Password control --

// Hash a password
func hashPw(pw string) []byte {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pw), HASH_COST)
	return hash
}

// Compare password and hash for validity.
func comparePw(pw string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
