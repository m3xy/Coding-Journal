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
	SUBROUTE_JOURNAL = "/supergroup"

	ENDPOINT_USERINFO = "/users"
	ENDPOINT_VALIDATE = "/validate"

	// general endpoints used in multiple sub-routes
	ENDPOINT_QUERY = "/query"
	ENDPOINT_DELETE = "/delete"
	ENDPOINT_EDIT = "/edit"
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
		AllowedOrigins: []string{"http://0.0.0.0:23409", "http://localhost:23409", "https://cs3099user11.host.cs.st-andrews.ac.uk"},
		AllowedHeaders: []string{"content-type", SECURITY_TOKEN_KEY, "BearerToken", "RefreshToken", "user"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "PUT"},
	})

	// Setup all routes.
	router.HandleFunc(ENDPOINT_VALIDATE, tokenValidation).Methods(http.MethodGet, http.MethodOptions)
	getAuthSubRoutes(router)        // Auth subroutes
	getJournalSubroute(router)      // Journal subroutes
	getUserSubroutes(router)        // Users subroutes
	getSubmissionsSubRoutes(router) // Submissions and files routes
	getFilesSubRoutes(router)

	// Setup HTTP server and shutdown signal notification
	return &http.Server{
		Addr:         PORT,
		Handler:      RequestLoggerMiddleware(c.Handler(router)),
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
