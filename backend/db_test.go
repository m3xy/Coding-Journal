package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"gorm.io/gorm/logger"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	TEST_LOG_PATH = "./test.log"
	TEST_PORT     = ":59213"
	LOCALHOST     = "http://localhost"
	TEST_DB       = "testdb"
	JSON_TAG_PW   = "password"
)

// -------------
// Dummy objects
// -------------

var testUsers []User = []User{
	{Email: "test.test@st-andrews.ac.uk", Password: "123456aB$", FirstName: "test",
		LastName: "test", PhoneNumber: "0574349206"},
	{Email: "john.doe@hello.com", Password: "dlbjDs2!", FirstName: "John",
		LastName: "Doe", Organization: "TestOrg"},
	{Email: "jane.doe@test.net", Password: "dlbjDs2!", FirstName: "Jane",
		LastName: "Doe"},
}

var testObjects []GlobalUser = []GlobalUser{
	{ID: "1", UserType: USERTYPE_REVIEWER_PUBLISHER},
	{ID: "2", UserType: USERTYPE_REVIEWER_PUBLISHER},
	{ID: "3", UserType: USERTYPE_REVIEWER_PUBLISHER},
	{ID: "4", UserType: USERTYPE_REVIEWER_PUBLISHER},
}

var wrongCredsUsers []User = []User{
	{Email: "test.nospec@hello.com", Password: "badN0Special", FirstName: "test", LastName: "nospec"},
	{Email: "test.nonum@hello.com", Password: "testNoNum!", FirstName: "test", LastName: "nonum"},
	{Email: "test.toosmall@hello.com", Password: "g0.Ku", FirstName: "test", LastName: "toosmall"},
	{Email: "test.wrongchars@hello.com", Password: "Tho/se]chars|ille\"gal", FirstName: "test", LastName: "wrongchars"},
	{Email: "test.nolowercase@hello.com", Password: "ALLCAP5!", FirstName: "test", LastName: "nolower"},
	{Email: "test.nouppercase@hello.com", Password: "nocap5!!", FirstName: "test", LastName: "noupper"},
}

var testLogger logger.Interface = logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
	SlowThreshold:             time.Second,
	LogLevel:                  logger.Error,
	IgnoreRecordNotFoundError: true,
	Colorful:                  true,
})

// -----------
// Utils
// -----------

// Initialise the database for testing.
func testInit() {
	gormDb, _ = gormInit(TEST_DB, testLogger)
	setup(gormDb, TEST_LOG_PATH)
	if err := gormDb.Transaction(gormClear); err != nil {
		fmt.Printf("Error occurred while clearing Db: %v", err)
	}
	// clears the filesystem
	if _, err := os.Stat(TEST_FILES_DIR); err == nil {
		os.RemoveAll(TEST_FILES_DIR)
	}
	if err := os.Mkdir(TEST_FILES_DIR, DIR_PERMISSIONS); err != nil {
		fmt.Printf("Error while clearing filesystem: %v", err)
	}
}

// Get a copy of a user object.
func (u *User) getCopy() User {
	return User{Email: u.Email, Password: u.Password, FirstName: u.FirstName,
		LastName: u.LastName, PhoneNumber: u.PhoneNumber, Organization: u.Organization}
}
func (u *GlobalUser) getCopy() GlobalUser {
	return GlobalUser{ID: u.ID, FullName: u.FullName}
}

// Get a copy of a user array.
func getUserCopies(uc []User) []User {
	res := make([]User, len(uc))
	for i, u := range uc {
		res[i] = u.getCopy()
	}
	return res
}
func getGlobalCopies(gc []User) []GlobalUser {
	res := make([]GlobalUser, len(gc))
	for i, u := range gc {
		res[i] = GlobalUser{User: u.getCopy(), UserType: USERTYPE_REVIEWER_PUBLISHER}
	}
	return res

}
func getObjectCopies(gc []GlobalUser) []GlobalUser {
	res := make([]GlobalUser, len(gc))
	for i, u := range gc {
		res[i] = u.getCopy()
	}
	return res
}

func testingServerSetup() *http.Server {
	router := mux.NewRouter()

	// Call authentication endpoints.
	router.HandleFunc(ENDPOINT_VALIDATE, tokenValidation).Methods(http.MethodGet)
	router.HandleFunc("/users/{"+getJsonTag(&GlobalUser{}, "ID")+"}", getUserProfile).Methods(http.MethodGet)

	// Setup testing HTTP server
	return &http.Server{
		Addr:    TEST_PORT,
		Handler: router,
	}
}

// Close database at the end of test.
func testEnd() {
	if err := gormDb.Transaction(gormClear); err != nil {
		fmt.Printf("Error occurred while clearing Db: %v", err)
	}
	getDB, _ := gormDb.DB()
	getDB.Close()
}

// Send a secure request in a JSON body from a given interface.
func sendJsonRequest(endpoint string, method string, data interface{}, port ...string) (*http.Response, error) {
	usedPort := TEST_PORT
	if len(port) > 0 {
		usedPort = port[0]
	}
	var req *http.Request
	if data != nil {
		jsonDat, _ := json.Marshal(data)
		req, _ = http.NewRequest(method, LOCALHOST+usedPort+endpoint, bytes.NewBuffer(jsonDat))
	} else {
		req, _ = http.NewRequest(method, LOCALHOST+usedPort+endpoint, nil)
	}
	return sendSecureRequest(gormDb, req, TEAM_ID)
}

// Set up authentication on a test server.
func testAuth(t *testing.T) {
	// Set up database.
	var err error
	gormDb, err = gormInit(TEST_DB, testLogger)
	if err != nil {
		fmt.Printf("Error opening database :%v", err)
	}
	if err := gormDb.Transaction(gormClear); err != nil {
		fmt.Printf("Error occured while clearing Db: %v", err)
	}

	// Set up logging to a local testing file.
	file, err := os.OpenFile(TEST_LOG_PATH, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Log file creation failure: %v! Exiting...", err)
	}
	log.SetOutput(file)
}

// -------------
// Tests
// -------------

// Test database initialisation
func TestDbInit(t *testing.T) {
	testDb, err := gormInit(dbname, testLogger)
	if err != nil {
		t.Error(err.Error())
	}
	getDB, _ := testDb.DB()
	getDB.Close()
}

// Test credential uniqueness with test database.
func TestIsUnique(t *testing.T) {
	testInit()

	// Test uniqueness in empty table
	t.Run("Unique elements", func(t *testing.T) {
		for i := 0; i < len(testObjects); i++ {
			assert.Truef(t, isUnique(gormDb, &GlobalUser{}, "ID", testObjects[i].ID),
				"ID %s Shouldn't exist in database!", testObjects[i].ID)
		}
	})

	// Add an element to table
	// Add test users to database
	trialObjects := getObjectCopies(testObjects)
	if err := gormDb.Create(&trialObjects).Error; err != nil {
		t.Errorf("Batch user creation error: %v", err)
		return
	}

	// Test uniquenes if element already exists in table.
	t.Run("Non-unique elements", func(t *testing.T) {
		for i := 0; i < len(testObjects); i++ {
			assert.Falsef(t, isUnique(gormDb, &GlobalUser{}, "ID", testObjects[i].ID),
				"Email %s should already be in database!", testObjects[i].ID)
		}
	})
	testEnd()
}
