package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
	"errors"
)

const (
	BASE64_CHARS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/"
	SECURITY_TOKEN_KEY = "X-FOREIGNJOURNAL-SECURITY-TOKEN"
	SECURITY_KEY_SIZE = 128
	LOG_FILE_PATH = "../cs3099-backend.log"
)
var CORSHeaders map[string]string = map[string]string {
	"Access-Control-Allow-Origin": "*",
	"Access-Control-Allow-Methods": "POST, GET, OPTIONS, PUT, DELETE",
	"Acess-Control-Allow-Headers": "X-Requested-With, " + SECURITY_TOKEN_KEY,
}

// Array of servers to connect to.
var serverArr []*Servers = []*Servers{
	{GroupNb: 23,
	Url: "https://cs3099user23.host.cs.st-andrews.ac.uk/api/v1/supergroup",
	Token: "0OTb5kV+qF9uD/bZ+kNp5vQ+O9PxznwtD9qDtVtQBHul4J+PENURYEQV0tayCISU"},
	{GroupNb: 5,
	Url: "cs3099user5.host.cs.st-andrews.ac.uk/api/v1/supergroup",
	Token: "LSVXQO1-w90P7XHYXdndSNPrUMUPQwiXzyJKdNqpgE5C6U0kZpZzKFk0eAiyHVwOI59M6GoyZSuDYyPNKe8ZYg"},
	{GroupNb: 13,
	Url: "https://cs3099user13.host.cs.st-andrews.ac.uk/api/v1/supergroup",
	Token: "0Wl/EtiV7N8g8yUR6UHOcLcystFy9SjxLEGO3uUD34Tkozr7+xAAZQxMQqOs2dUFUIjPFHuMmOWyjqAaqcDVIvM4AtxLCLADbaUvwlV/YmFSd9++HCrp76G8oaPcfzzcXN0q7T6yAie6thO4/zBN1nb2QFAfIRSWXj1E4DwRftc="},
	{GroupNb: 26,
	Url: "https://cs3099user26.host.cs.st-andrews.ac.uk/api/v1/supergroup",
	Token: "Mwjq2CmTcMhQovsBpOUBSCI20VSphI4o6nsaSs3yLeYklFQAKt1D5tklGLSa4svk0LJ8mKQ730YDk8Osme+KceIiJElEsQVH3NmEtU1ySqd0Lt+TUmsNf6ou3JAClcD1yUAbhosbVNnRMEHY0awK9wuJ2Vb7RnthWG4tgZcgQ+Q="},
	{GroupNb: 2,
	Url: "https://cs3099user2.host.cs.st-andrews.ac.uk/api/v1/supergroup",
	Token: "NWE0ZGExYTAxZTM2NjU4MjIxNTVmYzkyOTlkNGQ5OGFiMTFiMWI5NDEwY2RmNDhiODM2NGM5NGJhMDM0Mjc3N2E5NDMzNzQyZDM0NDcxNjk0NDU4NzdlYzljMjM3YTZhZDlmOWUxNGMwOGEwMTM2ZjI4MTI0YmM5MTRlZjliYmU0ZTg1NDc4OWY0NDI3YmFjZWM3MDBhMWU4OGNlZTIwZjM1NWFmZTlhZjFhZWEzNzA5ODE1MzVhMmUzZmMyYTE1ZjI3ZWQ3Mjc2ODM2NDcxYzA2ZTRjYjFhMzAzMDIwYWU2ODAxMWFlY2MwZWQ5MzE2NTg1ZTNkNmJmYjM5NjZkMQ"},
}

// Generate a new security key.
func getNewSecurityKey() string {
	return randStringBase64(int(time.Now().UnixNano()), SECURITY_KEY_SIZE)
}

// Generate a random base64 string.
func randStringBase64(seed int, n int) string {
	retStr := ""
	for i := 0; i < n; i++ {
		rand.Seed(int64(seed))
		randIndex := rand.Int() % 64
		retStr = fmt.Sprintf(retStr + "%c", BASE64_CHARS[randIndex])
		seed++
	}
	return retStr
}

// Check needed configuration setup before running server
func securityCheck() error {
	// Set log file logging.
	file, err := os.OpenFile(LOG_FILE_PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Log file creation failure: %v! Exitting...", err)
		return err
	}
	log.SetOutput(file)

	// Check security token existence before running.
	if checkUnique(TABLE_SERVERS, getDbTag(&Servers{}, "GroupNb"), TEAM_ID) {
		log.Println("Server token not set! Setting up...")
		securityToken := getNewSecurityKey()
		_, err := db.Exec(fmt.Sprintf("INSERT INTO %s VALUES (?, ?, ?);", TABLE_SERVERS),
			TEAM_ID, securityToken, allowedOrigins[0])
		if err != nil {
			log.Fatalf("Critical token failure! %v\n", err)
			return err
		} else {
			log.Println("Security token successfully stored in database.")
			log.Printf("Store this security token: %s\n", securityToken)
		}
	}

	// Check for filesystem existence.
	if _, err := os.Stat(FILESYSTEM_ROOT); os.IsNotExist(err)  {
		log.Printf("Filesystem not setup up! Setting it up at %s",
		FILESYSTEM_ROOT)
		os.MkdirAll(FILESYSTEM_ROOT, DIR_PERMISSIONS)
	}

	// Set server tokens for all servers in organization.
	for _, server := range serverArr {
		stmt := fmt.Sprintf(SELECT_ROW, getDbTag(&Servers{}, "Token"),TABLE_SERVERS,
	getDbTag(&Servers{}, "GroupNb"))
		var token string
		err := db.QueryRow(stmt, server.GroupNb).Scan(&token)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Server no. %d not set up! Setting it up...", server.GroupNb)
				_, err = db.Exec(
					fmt.Sprintf("INSERT INTO %s VALUES (?, ?, ?);", TABLE_SERVERS),
					server.GroupNb, server.Token, server.Url)
				if err != nil {
					log.Fatalf("FATAL - Server insertion error: %v. Exiting...", err)
					return err
				}
			} else {
				log.Fatalf("FATAL - Database error: %v! Exiting...", err)
				return err
			}
		}
	}
	return nil
}

// Send a given request using needed authentication.
func sendSecureRequest(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("Request nil!")
	}
	// Fetch valid security token from database.
	storedToken := ""
	err := db.QueryRow(fmt.Sprintf(
		SELECT_ROW, getDbTag(&Servers{}, "Token"), TABLE_SERVERS,
		getDbTag(&Servers{}, "GroupNb")), TEAM_ID).
		Scan(&storedToken)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req.Header.Set(SECURITY_TOKEN_KEY, storedToken)
	return client.Do(req)
}

// Setup a response when OPTIONS is received.
func setupResponse(w *http.ResponseWriter, r *http.Request) {
	for key, val := range CORSHeaders {
		(*w).Header().Set(key, val)
	}
}

// Validate the given security token's authenticity.
func validateToken(token string) bool {
	// Query token from servers table.
	storedToken := ""
	err := db.QueryRow(fmt.Sprintf(
		SELECT_ROW, getDbTag(&Servers{}, "Token"), TABLE_SERVERS,
		getDbTag(&Servers{}, "GroupNb")), TEAM_ID).
		Scan(&storedToken)
	if err != nil || storedToken != token {
		return false
	} else  {
		return true
	}
}

// Validate if given security token works.
// Params:
// 	Header: securityToken
// Return:
//  200: Success - security token valid.
//  401: Failure - security token invalid.
func tokenValidation(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(SECURITY_TOKEN_KEY)
	if !validateToken(token) {
		w.WriteHeader(http.StatusUnauthorized)
	}
}
