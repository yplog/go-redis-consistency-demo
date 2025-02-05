package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// Redis Connections
var mainRedisClient *redis.Client
var replicaRedisClient *redis.Client

// Mode
var consistencyMode string

// Redis key constants
const CounterKey = "counter"

func main() {
	flag.StringVar(&consistencyMode, "mode", "eventual", "Set consistency mode: 'strong' or 'eventual'")
	flag.Parse()

	// Main Redis Client for writing data
	mainRedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Replica Redis Client for reading data
	replicaRedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})

	// HTTP Endpoints
	http.HandleFunc("/increment", handleIncrement)
	http.HandleFunc("/get", handleGet)

	log.Printf("Server running on :8080 (Mode: %s)", consistencyMode)
	http.ListenAndServe(":8080", nil)
}

// Increment the counter always write to main Redis
func handleIncrement(w http.ResponseWriter, r *http.Request) {
	if consistencyMode == "strong" {
		// Wait for replication to complete with WAIT command
		replicas, err := mainRedisClient.Wait(ctx, 1, 5000).Result()
		if err != nil {
			http.Error(w, "Replication wait failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if replicas < 1 {
			http.Error(w, "Could not replicate to enough replicas", http.StatusInternalServerError)
			return
		}
	}

	newValue, err := mainRedisClient.Incr(ctx, CounterKey).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Counter updated to: %d", newValue)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	if consistencyMode == "strong" {
		// STRONG MODE: Read from slave but verify with the latest written value
		for i := 0; i < 5; i++ { // Try 5 times
			slaveVal, err := replicaRedisClient.Get(ctx, CounterKey).Result()
			if err == redis.Nil {
				slaveVal = "0"
			} else if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Get the latest value from the main
			masterVal, err := mainRedisClient.Get(ctx, CounterKey).Result()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// If the main and replica return the same value, strong consistency is achieved
			if slaveVal == masterVal {
				fmt.Fprintf(w, "Counter Value (Strong Mode): %s", slaveVal)
				return
			}

			// If the slave is not updated, wait a bit and try again
			time.Sleep(200 * time.Millisecond)
		}

		// If still no match, read from the main
		masterVal, _ := mainRedisClient.Get(ctx, CounterKey).Result()
		fmt.Fprintf(w, "Counter Value (Strong Mode, Forced Master Read): %s", masterVal)
		return
	}

	// EVENTUAL MODE: Directly read from the replica
	slaveVal, err := replicaRedisClient.Get(ctx, CounterKey).Result()
	if err == redis.Nil {
		fmt.Fprintf(w, "Counter Value (Eventual Mode): %d", 0)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Counter Value (Eventual Mode): %s", slaveVal)
}
