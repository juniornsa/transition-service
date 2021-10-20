package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"upworkfixmux/handler"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	h := &handler.UserHandler{
		Ch: make(chan string, 1),
	}

	router := mux.NewRouter()
	// router.Use(middleware.LoginMiddleware)
	router.HandleFunc("/", h.Get).Methods(http.MethodGet)
	router.HandleFunc("/tag", h.Get).Methods(http.MethodGet)
	router.HandleFunc("/tag/{jira_id}", h.Release).Methods(http.MethodGet)
	router.HandleFunc("/slack", h.Slack).Methods(http.MethodPost)

	log.Println("Starting server on port: 8082")
	log.Fatal(http.ListenAndServe(":8082", router))
}
