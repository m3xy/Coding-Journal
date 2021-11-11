package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	BASE64_CHARS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/"
	SECURITY_TOKEN_KEY = "X-FOREIGNJOURNAL-SECURITY-TOKEN"
	SECURITY_KEY_SIZE = 128
)

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
			return nil
		}
	} else {
		return nil
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
