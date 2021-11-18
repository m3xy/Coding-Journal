package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const (
	// Constants for database connection.
	host     = "127.0.0.1"
	port     = 3307
	user     = "root"
	protocol = "tcp"
	password = "secret"
	dbname   = "mydb"

	BACKEND_ADDRESS = "http://localhost:3333/"
	PORT            = ":3333"
	ADDRESS_KEY     = "BACKEND_ADDRESS"
	ENV_DIR         = "../frontend/.env"

	// end points for URLs
	ENDPOINT_LOGIN        = "/login"
	ENDPOINT_LOGIN_GLOBAL = "/glogin"
	ENDPOINT_SIGNUP       = "/register"
	ENDPOINT_ALL_PROJECTS = "/projects"
	ENDPOINT_PROJECT      = "/project"
	ENDPOINT_FILE         = "/project/file"
	ENDPOINT_NEWFILE      = "/upload"
	ENDPOINT_USERINFO     = "/users"
	ENDPOINT_NEWCOMMENT   = "/project/file/newcomment"
	ENDPOINT_VALIDATE     = "/validate"
)

// Environment variables setter map.
var dotenvMap map[string]string = map[string]string{}

func main() {
	srv := setupCORSsrv()

	// Initialise database with production credentials.
	dbInit(user, password, protocol, host, port, dbname)
	setup()

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Run server in goroutine to avoid blocking call.
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
	// headersOk := handlers.AllowedHeaders(allowedHeaders)
	// originsOk := handlers.AllowedOrigins(allowedOrigins)
	// methodsOk := handlers.AllowedMethods(allowedMethods)

	// Setup all routes.
	router.HandleFunc(ENDPOINT_LOGIN, logIn).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_LOGIN_GLOBAL, logInGlobal).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_SIGNUP, signUp).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_ALL_PROJECTS, getAllProjects).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_PROJECT, getProject).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_FILE, getFile).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_NEWCOMMENT, uploadUserComment).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_NEWFILE, uploadSingleFile).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_VALIDATE, tokenValidation).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/users/{"+getJsonTag(&Credentials{}, "Id")+"}", getUserProfile).Methods(http.MethodGet, http.MethodOptions)

	// Setup HTTP server and shutdown signal notification
	return &http.Server{
		Addr: PORT,
		// Handler: handlers.CORS(originsOk, headersOk, methodsOk)(router),
		Handler: router,
	}
}

// Set up the server and it's dependencies.
func setup() error {
	// Set log file logging.
	var err error = nil
	file, err := os.OpenFile(LOG_FILE_PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Log file creation failure: %v! Exitting...", err)
		goto RETURN
	}
	log.SetOutput(file)

	// Check security token existence before running.
	err = securityCheck()
	if err != nil {
		goto RETURN
	}

	// Check for filesystem existence.
	if _, err = os.Stat(FILESYSTEM_ROOT); os.IsNotExist(err) {
		log.Printf("Filesystem not setup up! Setting it up at %s",
			FILESYSTEM_ROOT)
		err = os.MkdirAll(FILESYSTEM_ROOT, DIR_PERMISSIONS)
		if err != nil {
			log.Fatalf("Filesystem setup error: %v\n", err)
			goto RETURN
		}
	}

	// Set foreign servers.
	if err = setForeignServers(); err != nil {
		log.Fatalf("FATAL - Foreign server set up error: %v\n", err)
		goto RETURN
	}

	// Write needed environment variables to dotenv file.
	dotenvMap[ADDRESS_KEY] = BACKEND_ADDRESS
	err = godotenv.Write(dotenvMap, ENV_DIR)

RETURN:
	return err
}
