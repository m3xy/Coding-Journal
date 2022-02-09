package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	// Constants for database connection.
	user     = "root"
	protocol = "tcp"
	password = "secret"
	dbname   = "mydb"

	BACKEND_ADDRESS = "http://localhost:3333"
	PORT            = ":3333"
	ADDRESS_KEY     = "BACKEND_ADDRESS"

	// end points for URLs
	SUBROUTE_AUTH            = "/auth"
	ENDPOINT_LOGIN           = "/login"
	ENDPOINT_SIGNUP          = "/register"
	ENDPOINT_ALL_SUBMISSIONS = "/submissions"
	ENDPOINT_SUBMISSION      = "/submission"
	ENDPOINT_FILE            = "/submission/file"
	ENDPOINT_NEWFILE         = "/upload"
	ENDPOINT_USERINFO        = "/users"
	ENDPOINT_NEWCOMMENT      = "/submission/file/newcomment"
	ENDPOINT_VALIDATE        = "/validate"
)

var prodLogger logger.Interface = logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
	SlowThreshold: time.Second,
	LogLevel:      logger.Error,
})

func main() {

	// Initialise database with production credentials.
	var err error
	if gormDb, err = gormInit(dbname, prodLogger); err != nil {
		return
	}
	setup(gormDb, os.Getenv("LOG_PATH"))

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Run server in goroutine to avoid blocking call.
	srv := setupCORSsrv()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v\n", err)
		}
	}()
	log.Printf("Server started.\n")

	<-done // Wait for termination signal to be received before ending program.
	log.Printf("Server stopped.\n")

	// Gracefully shut down server by shutting down all idling connections after a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown failed: %v\n", err)
	}
	log.Printf("Server shut down properly.\n")
}

// Setup CORS-compatible server.
func setupCORSsrv() *http.Server {
	// Set up login and signup functions
	router := mux.NewRouter()

	// sets up handler for CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:23409", "https://cs3099user11.host.cs.st-andrews.ac.uk"},
		AllowedHeaders: []string{"content-type", SECURITY_TOKEN_KEY, "bearer_token", "refresh_token", "user"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "PUT"},
	})

	// Set up middleware
	router.Use(authenticationMiddleWare)

	// Setup all routes.
	router.HandleFunc(ENDPOINT_LOGIN, logIn).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_ALL_SUBMISSIONS, getAllSubmissions).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc(ENDPOINT_SUBMISSION, sendSubmission).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc(ENDPOINT_FILE, getFile).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc(ENDPOINT_NEWCOMMENT, uploadUserComment).Methods(http.MethodPost, http.MethodOptions)
	// router.HandleFunc(ENDPOINT_NEWFILE, uploadSingleFile).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_VALIDATE, tokenValidation).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/users/{"+getJsonTag(&GlobalUser{}, "ID")+"}", getUserProfile).Methods(http.MethodGet, http.MethodOptions)

	// Auth subroutes
	auth := router.PathPrefix(SUBROUTE_AUTH).Subrouter()
	getAuthSubRoutes(auth)

	// Setup HTTP server and shutdown signal notification
	return &http.Server{
		Addr:         PORT,
		Handler:      c.Handler(router),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

// Set up the server and it's dependencies.
func setup(db *gorm.DB, logpath string) error {
	// Set log file logging.
	var err error = nil
	file, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Log file creation failure: %v! Exitting...", err)
		goto RETURN
	}
	log.SetOutput(file)

	// Check security token existence before running.
	err = securityCheck(db)
	if err != nil {
		goto RETURN
	}

	// Set validation functions
	validator.SetValidationFunc("ispw", ispw)
	validator.SetValidationFunc("isemail", isemail)

	// Check for filesystem existence.
	/* if _, err = os.Stat(FILESYSTEM_ROOT); os.IsNotExist(err) {
		log.Printf("Filesystem not setup up! Setting it up at %s",
			FILESYSTEM_ROOT)
		err = os.MkdirAll(FILESYSTEM_ROOT, DIR_PERMISSIONS)
		if err != nil {
			log.Fatalf("Filesystem setup error: %v\n", err)
			goto RETURN
		}
	} */

	// Set foreign servers.
	if err = setForeignServers(db); err != nil {
		log.Fatalf("FATAL - Foreign server set up error: %v\n", err)
		goto RETURN
	}

RETURN:
	return err
}
