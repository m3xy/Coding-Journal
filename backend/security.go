package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

const (
	BASE64_CHARS       = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/"
	SECURITY_TOKEN_ENV = "BACKEND_TOKEN"
	SECURITY_KEY_SIZE  = 128
	LOG_FILE_PATH      = "./cs3099-backend.log"
)

// Array of servers to connect to.
var serverArr []Server = []Server{
	{GroupNumber: 23,
		Url:   "https://cs3099user23.host.cs.st-andrews.ac.uk/api/v1/supergroup",
		Token: "0OTb5kV+qF9uD/bZ+kNp5vQ+O9PxznwtD9qDtVtQBHul4J+PENURYEQV0tayCISU"},
	{GroupNumber: 5,
		Url:   "cs3099user05.host.cs.st-andrews.ac.uk/api/v1/supergroup",
		Token: "LSVXQO1-w90P7XHYXdndSNPrUMUPQwiXzyJKdNqpgE5C6U0kZpZzKFk0eAiyHVwOI59M6GoyZSuDYyPNKe8ZYg"},
	{GroupNumber: 13,
		Url:   "https://cs3099user13.host.cs.st-andrews.ac.uk/api/v1/supergroup",
		Token: "0Wl/EtiV7N8g8yUR6UHOcLcystFy9SjxLEGO3uUD34Tkozr7+xAAZQxMQqOs2dUFUIjPFHuMmOWyjqAaqcDVIvM4AtxLCLADbaUvwlV/YmFSd9++HCrp76G8oaPcfzzcXN0q7T6yAie6thO4/zBN1nb2QFAfIRSWXj1E4DwRftc="},
	{GroupNumber: 26,
		Url:   "https://cs3099user26.host.cs.st-andrews.ac.uk/api/v1/supergroup",
		Token: "Mwjq2CmTcMhQovsBpOUBSCI20VSphI4o6nsaSs3yLeYklFQAKt1D5tklGLSa4svk0LJ8mKQ730YDk8Osme+KceIiJElEsQVH3NmEtU1ySqd0Lt+TUmsNf6ou3JAClcD1yUAbhosbVNnRMEHY0awK9wuJ2Vb7RnthWG4tgZcgQ+Q="},
	{GroupNumber: 2,
		Url:   "https://cs3099user02.host.cs.st-andrews.ac.uk/api/v1/supergroup",
		Token: "NWE0ZGExYTAxZTM2NjU4MjIxNTVmYzkyOTlkNGQ5OGFiMTFiMWI5NDEwY2RmNDhiODM2NGM5NGJhMDM0Mjc3N2E5NDMzNzQyZDM0NDcxNjk0NDU4NzdlYzljMjM3YTZhZDlmOWUxNGMwOGEwMTM2ZjI4MTI0YmM5MTRlZjliYmU0ZTg1NDc4OWY0NDI3YmFjZWM3MDBhMWU4OGNlZTIwZjM1NWFmZTlhZjFhZWEzNzA5ODE1MzVhMmUzZmMyYTE1ZjI3ZWQ3Mjc2ODM2NDcxYzA2ZTRjYjFhMzAzMDIwYWU2ODAxMWFlY2MwZWQ5MzE2NTg1ZTNkNmJmYjM5NjZkMQ"},
	{GroupNumber: 20,
		Url:   "https://cs3099user20.host.cs.st-andrews.ac.uk/api/v1/supergroup",
		Token: "cs3099user20ThisIsASecretTokenPlzDontShare:)"},
}

// Get security key from database.
func getDbSecurityKey(db *gorm.DB) (string, error) {
	// Get security key from database
	var server Server
	if res := db.Select("Token").Limit(1).Find(&server, TEAM_ID); res.Error != nil {
		return "", res.Error
	} else if res.RowsAffected == 0 {
		return "", nil
	} else {
		return server.Token, nil
	}
}

// Check security key existence, and set it if it doesn't exist.
func securityCheck(db *gorm.DB) error {
	// Check security token existence before running.
	token, err := getDbSecurityKey(db)
	if err != nil {
		log.Fatalf("Can't fetch security token: %v", err)
		return err
	} else if token == "" {
		log.Println("[WARN] Server token not set! Setting up...")
		myEnv, err := godotenv.Read("secrets.env")
		if err != nil {
			log.Fatal("No token has been set!")
			return err
		}
		securityToken := myEnv[SECURITY_TOKEN_ENV]
		if err := db.Create(&Server{GroupNumber: TEAM_ID, Token: securityToken, Url: BACKEND_ADDRESS}).Error; err != nil {
			log.Fatalf("Critical token failure! %v\n", err)
			return err
		} else {
			log.Println("[INFO] Security token successfully stored in the database.")
		}
	}
	return nil
}

// Add foreign servers to database if not added yet.
func setForeignServers(db *gorm.DB) error {
	var err error = nil
	// Set server tokens for all servers in organization.
	for _, server := range serverArr {
		var result Server
		res := db.Where(&server).Limit(1).Find(&result)
		if res.RowsAffected == 0 || res.Error != nil {
			if res.RowsAffected == 0 {
				log.Printf("Server no. %d not set up! Setting it up...", server.GroupNumber)
				db.Create(&server)
			}
		}
	}
	return err
}

// Send a given request using needed authentication.
func sendSecureRequest(db *gorm.DB, req *http.Request, groupNb int) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("Request nil!")
	}
	// Fetch valid security token from database.
	var storedToken string
	if err := db.Model(&Server{}).Select("token").Find(&storedToken, groupNb).Error; err != nil {
		return nil, err
	}
	client := &http.Client{}
	req.Header.Set(SECURITY_TOKEN_KEY, storedToken)
	return client.Do(req)
}

// Validate the given security token's authenticity.
func validateSecurityKey(db *gorm.DB, token string) bool {
	// Query token from servers table.
	var exists bool
	if err := db.Model(&Server{}).Select("count(*) > 0").Where("token = ?", token).Limit(1).Find(&exists).Error; err != nil {
		return false
	} else {
		return exists
	}
}
