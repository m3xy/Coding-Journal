package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Set up login and signup functions
	router := mux.NewRouter()
	router.HandleFunc("/login", logIn)
	router.HandleFunc("/signup", signUp)
	// Initialise database
	dbInit()

	log.Fatal(http.ListenAndServe(":8080", router)) // Main program runner.

}
