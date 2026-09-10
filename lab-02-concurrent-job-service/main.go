package main

import (
	"log"
	"net/http"
)

func main() {
	server := NewServer(100)

	server.startWorkers(5)

	http.HandleFunc("/jobs", server.handleJobs)

	log.Println("server listening on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
