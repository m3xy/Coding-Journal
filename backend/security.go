package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

const (
	BASE64_CHARS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/"
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
