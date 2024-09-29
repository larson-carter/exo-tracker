package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load("/Users/larsoncarter/Documents/GIT-REPOS/exo-tracker/peer/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	peerID := os.Getenv("PEER_ID")
	peerIP := os.Getenv("PEER_IP")
	peerPort := os.Getenv("PEER_PORT")
	trackerURL := os.Getenv("TRACKER_URL")

	peer := map[string]string{
		"id":   peerID,
		"ip":   peerIP,
		"port": peerPort,
	}

	peerJSON, _ := json.Marshal(peer)

	res, err := http.Post(trackerURL+"/register", "application/json", bytes.NewBuffer(peerJSON))
	if err != nil {
		log.Fatalf("Failed to register with tracker: %v\n", err)
	}
	defer res.Body.Close()

	log.Printf("Peer %s registered with tracker at %s\n", peerID, trackerURL)
	
}
