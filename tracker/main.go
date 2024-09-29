package main

import (
	"exo-tracker/tracker/api" // Import the API handlers
	"log"
	"net/http"
	"time"
)

const heartbeatTimeout = 30 * time.Second

func main() {
	go api.MonitorPeers(heartbeatTimeout) // Start peer monitoring

	http.HandleFunc("/register", api.RegisterPeer)
	http.HandleFunc("/heartbeat", api.HeartbeatPeer) // New heartbeat endpoint
	http.HandleFunc("/peers", api.GetPeers)
	http.HandleFunc("/deregister", api.DeregisterPeer)

	log.Println("Tracker service running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
