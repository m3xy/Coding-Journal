package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	// Constants for database connection.
	host     = "127.0.0.1"
	port     = 3307
	user     = "root"
	protocol = "tcp"
	password = "secret"
	dbname   = "mydb"

	// end points for URLs
	ENDPOINT_LOGIN = "/login"
	ENDPOINT_SIGNUP = "/register"
	ENDPOINT_ALL_PROJECTS = "/projects"
	ENDPOINT_PROJECT = "/project"
	ENDPOINT_FILE = "/project/file"
	ENDPOINT_NEWFILE = "/upload"
	ENDPOINT_NEWCOMMENT = "/project/file/newcomment"
	ENDPOINT_VALIDATE = "/validate"
)

var allowedOrigins = []string{"http://localhost:3333",
	"http://localhost:23409", "http://localhost:10533"}

func main() {
	srv := setupCORSsrv()

	// Initialise database with production credentials.
	dbInit(user, password, protocol, host, port, dbname)
	securityCheck()
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Run server in goroutine to avoid blocking call.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v\n", err)
		}
	}()
	log.Printf("Server started at %s\n", time.Now().String())

	<-done // Wait for termination signal to be received before ending program.
	log.Printf("Server stopped at %s\n", time.Now().String())

	// Gracefully shut down server by shutting down all idling connections after a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown failed: %v\n", err)
	}
	log.Printf("Server shut down properly at %s\n", time.Now().String())
}

// Setup CORS-compatible server.
func setupCORSsrv() *http.Server {
	// Set up login and signup functions
	router := mux.NewRouter()
	router.HandleFunc(ENDPOINT_LOGIN, logIn)
	router.HandleFunc(ENDPOINT_SIGNUP, signUp)
	router.HandleFunc(ENDPOINT_ALL_PROJECTS, getAllProjects)
	router.HandleFunc(ENDPOINT_PROJECT, getProject)
	router.HandleFunc(ENDPOINT_FILE, getFile)
	router.HandleFunc(ENDPOINT_NEWCOMMENT, uploadUserComment)
	router.HandleFunc(ENDPOINT_NEWFILE, uploadSingleFile)
	router.HandleFunc(ENDPOINT_VALIDATE, tokenValidation)
	router.HandleFunc("/users/{"+getJsonTag(&Credentials{},"Id")+"}", getUserProfile)

	// sets up handler for CORS
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", SECURITY_TOKEN_KEY})
	originsOk := handlers.AllowedOrigins(allowedOrigins)
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	// Setup HTTP server and shutdown signal notification
	return &http.Server{
		Addr:    ":3333",
		Handler: handlers.CORS(originsOk, headersOk, methodsOk)(router),
	}
}
