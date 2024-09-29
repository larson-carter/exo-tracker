package main

import (
	"exo-tracker/tracker/api" // Import the API handlers
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/register", api.RegisterPeer) // Move handlers to /api
	http.HandleFunc("/peers", api.GetPeers)
	http.HandleFunc("/deregister", api.DeregisterPeer)

	log.Println("Tracker service running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
