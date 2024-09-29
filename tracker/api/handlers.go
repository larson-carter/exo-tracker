package api

import (
	"encoding/json"
	"exo-tracker/common" // Shared data models
	"log"
	"net/http"
)

var peers = make(map[string]models.Peer)

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
