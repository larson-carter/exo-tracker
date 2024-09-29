package main

import (
	"bytes"
	"encoding/json"
	"exo-tracker/peer/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const heartbeatInterval = 10 * time.Second // Interval for sending heartbeats

func main() {
	//err := godotenv.Load("/Users/larsoncarter/Documents/GIT-REPOS/exo-tracker/peer/.env.peer2")
	err := godotenv.Load("/Users/larsoncarter/Documents/GIT-REPOS/exo-tracker/peer/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get environment variables from the .env file
	peerID := os.Getenv("PEER_ID")
	peerIP := os.Getenv("PEER_IP")
	peerPort := os.Getenv("PEER_PORT")
	trackerURL := os.Getenv("TRACKER_URL")

	// Register the peer with the tracker
	registerPeer(peerID, peerIP, peerPort, trackerURL)

	// Periodically send heartbeats
	go func() {
		for {
			sendHeartbeat(peerID, trackerURL)
			time.Sleep(heartbeatInterval) // Wait before sending the next heartbeat
		}
	}()

	// Periodically fetch and send messages to other peers
	go func() {
		for {
			peers, err := fetchPeers(trackerURL)
			if err != nil {
				log.Printf("Failed to fetch peers: %v\n", err)
				continue
			}

			for _, peer := range peers {
				if peer.ID != peerID { // Don't send to yourself
					err := sendMessageToPeer(peer, "Hello from "+peerID)
					if err != nil {
						log.Printf("Failed to send message to %s: %v\n", peer.ID, err)
					} else {
						log.Printf("Sent message to %s\n", peer.ID)
					}
				}
			}
			time.Sleep(10 * time.Second) // Send messages every 10 seconds
		}
	}()

	// Handle incoming messages
	http.HandleFunc("/message", handleMessage)

	// Start the peer service
	log.Printf("Peer service %s running on port %s...", peerID, peerPort)
	if err := http.ListenAndServe(":"+peerPort, nil); err != nil {
		log.Fatal(err)
	}
}

// Register the peer with the tracker
func registerPeer(peerID, peerIP, peerPort, trackerURL string) {
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

// Send heartbeats to the tracker
func sendHeartbeat(peerID, trackerURL string) {
	heartbeatData := map[string]string{
		"id": peerID,
	}

	heartbeatJSON, _ := json.Marshal(heartbeatData)
	res, err := http.Post(trackerURL+"/heartbeat", "application/json", bytes.NewBuffer(heartbeatJSON))
	if err != nil {
		log.Printf("Failed to send heartbeat for peer %s: %v\n", peerID, err)
		return
	}
	defer res.Body.Close()

	log.Printf("Sent heartbeat for peer %s\n", peerID)
}

// Fetch peers from the tracker
func fetchPeers(trackerURL string) ([]models.Peer, error) {
	res, err := http.Get(trackerURL + "/peers")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var peerList []models.Peer
	err = json.NewDecoder(res.Body).Decode(&peerList)
	if err != nil {
		return nil, err
	}

	return peerList, nil
}

// Send message to a peer
func sendMessageToPeer(peer models.Peer, message string) error {
	url := "http://" + peer.IP + ":" + peer.Port + "/message"
	reqBody, _ := json.Marshal(map[string]string{
		"message": message,
	})

	res, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

// Handle incoming messages
func handleMessage(w http.ResponseWriter, r *http.Request) {
	var message map[string]string
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	log.Printf("Received message: %s\n", message["message"])
	w.WriteHeader(http.StatusOK)
}
