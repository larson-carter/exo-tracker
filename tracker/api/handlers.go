package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"exo-tracker/common" // Shared data models
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	mu             sync.Mutex
	rdb            *redis.Client
	db             *sql.DB
	ctx            = context.Background()
	peerTimestamps = make(map[string]time.Time) // Track last heartbeat
)

func init() {
	err := godotenv.Load("/Users/larsoncarter/Documents/GIT-REPOS/exo-tracker/tracker/api/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"), // Redis address from .env
	})

	// Initialize PostgreSQL connection
	connStr := os.Getenv("POSTGRES_URL")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v\n", err)
	}

	log.Println("Connected to PostgreSQL and Redis.")
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

	mu.Lock()
	defer mu.Unlock()

	// Insert peer into PostgreSQL
	_, err := db.Exec("INSERT INTO peers (id, ip, port) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET ip = $2, port = $3, last_heartbeat = NOW()", peer.ID, peer.IP, peer.Port)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Failed to insert peer: %v\n", err)
		return
	}

	// Cache peer in Redis
	err = rdb.Set(ctx, peer.ID, peer.IP+":"+peer.Port, 0).Err()
	if err != nil {
		http.Error(w, "Redis error", http.StatusInternalServerError)
		log.Printf("Failed to cache peer in Redis: %v\n", err)
		return
	}

	// Update heartbeat
	peerTimestamps[peer.ID] = time.Now()

	log.Printf("Registered new peer: %+v\n", peer)
	w.WriteHeader(http.StatusCreated)
}

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
	if _, exists := peerTimestamps[peer.ID]; exists {
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
				deletePeer(peerID)
			}
		}
		mu.Unlock()
	}
}

// Helper function to remove peer from both PostgreSQL and Redis
func deletePeer(peerID string) {
	_, err := db.Exec("DELETE FROM peers WHERE id = $1", peerID)
	if err != nil {
		log.Printf("Failed to delete peer from PostgreSQL: %v\n", err)
	}

	err = rdb.Del(ctx, peerID).Err()
	if err != nil {
		log.Printf("Failed to delete peer from Redis: %v\n", err)
	}

	delete(peerTimestamps, peerID)
}

// GetPeers handler
func GetPeers(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Get all peers from PostgreSQL
	rows, err := db.Query("SELECT id, ip, port FROM peers")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Failed to fetch peers: %v\n", err)
		return
	}
	defer rows.Close()

	var peerList []models.Peer
	for rows.Next() {
		var peer models.Peer
		if err := rows.Scan(&peer.ID, &peer.IP, &peer.Port); err != nil {
			log.Printf("Failed to scan peer: %v\n", err)
			continue
		}
		peerList = append(peerList, peer)
	}

	// Return peers as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(peerList); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// DeregisterPeer handler
func DeregisterPeer(w http.ResponseWriter, r *http.Request) {
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
	defer mu.Unlock()

	// Remove peer from PostgreSQL and Redis
	deletePeer(peer.ID)

	log.Printf("Deregistered peer: %s\n", peer.ID)
	w.WriteHeader(http.StatusOK)
}
