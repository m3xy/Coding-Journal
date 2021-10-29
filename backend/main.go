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
)

const (
	// Constants for database connection.
	host     = "127.0.0.1"
	port     = 3306
	user     = "root"
	protocol = "tcp"
	password = "secret"
	dbname   = "mydb"
)

func main() {
	// Set up login and signup functions
	router := mux.NewRouter()
	router.HandleFunc("/login", logIn)
	router.HandleFunc("/signup", signUp)

	// Initialise database with production credentials.
	dbInit(user, password, protocol, host, port, dbname)
	// Setup HTTP server and shutdown signal notification
	srv := &http.Server {
		Addr: ":8080",
		Handler: router,
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer func() {
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown failed: %v\n", err)
	}
	log.Printf("Server shut down properly at %s\n", time.Now().String())
}
