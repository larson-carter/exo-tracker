package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Peer struct {
	ID   string `json:"id"`
	IP   string `json:"ip"`
	Port string `json:"port"`
}

var peers = make(map[string]Peer)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the exo tracker!")
	})

	http.HandleFunc("/register", registerPeer)
	http.HandleFunc("/peers", getPeers)
	http.HandleFunc("/deregister", deregisterPeer)

	log.Println("Tracker running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func registerPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var peer Peer
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	peers[peer.ID] = peer
	log.Printf("Registered new peer: %+v\n", peer)
	w.WriteHeader(http.StatusCreated)
}

func getPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var peerList []Peer
	for _, peer := range peers {
		peerList = append(peerList, peer)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(peerList); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func deregisterPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var peer Peer
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	delete(peers, peer.ID)
	log.Printf("Deregistered peer: %s\n", peer.ID)
	w.WriteHeader(http.StatusOK)
}
