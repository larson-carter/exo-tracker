package api

import (
	"encoding/json"
	"exo-tracker/common" // Shared data models
	"log"
	"net/http"
	"sync"
	"time"
)

var peers = make(map[string]models.Peer)
var peerTimestamps = make(map[string]time.Time) // Track last heartbeat
var mu sync.Mutex                               // Mutex to protect access to maps

// HeartbeatPeer handler to receive heartbeats
func HeartbeatPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var peer models.Peer
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	if _, exists := peers[peer.ID]; exists {
		peerTimestamps[peer.ID] = time.Now() // Update heartbeat timestamp
		log.Printf("Received heartbeat from peer: %s\n", peer.ID)
	}
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

// MonitorPeers periodically checks for inactive peers
func MonitorPeers(timeout time.Duration) {
	for {
		time.Sleep(10 * time.Second) // Run every 10 seconds

		mu.Lock()
		for peerID, lastSeen := range peerTimestamps {
			if time.Since(lastSeen) > timeout {
				log.Printf("Peer %s timed out, removing it\n", peerID)
				delete(peers, peerID)
				delete(peerTimestamps, peerID)
			}
		}
		mu.Unlock()
	}
}

// RegisterPeer handler
func RegisterPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var peer models.Peer
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	peers[peer.ID] = peer
	log.Printf("Registered new peer: %+v\n", peer)
	w.WriteHeader(http.StatusCreated)
}

// GetPeers handler
func GetPeers(w http.ResponseWriter, r *http.Request) {
	var peerList []models.Peer
	for _, peer := range peers {
		peerList = append(peerList, peer)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(peerList); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// DeregisterPeer handler
func DeregisterPeer(w http.ResponseWriter, r *http.Request) {
	var peer models.Peer
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	delete(peers, peer.ID)
	log.Printf("Deregistered peer: %s\n", peer.ID)
	w.WriteHeader(http.StatusOK)
}
